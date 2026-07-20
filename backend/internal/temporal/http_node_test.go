package temporal

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

func TestValidateHTTPNodeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid",
			config: map[string]interface{}{
				"method":  "POST",
				"url":     "https://api.example.com/items",
				"headers": map[string]interface{}{"Content-Type": "application/json"},
				"body":    `{"ok":true}`,
			},
		},
		{name: "missing method", config: map[string]interface{}{"url": "https://api.example.com"}, wantErr: true},
		{name: "unsupported method", config: map[string]interface{}{"method": "TRACE", "url": "https://api.example.com"}, wantErr: true},
		{name: "relative URL", config: map[string]interface{}{"method": "GET", "url": "/private"}, wantErr: true},
		{name: "non HTTP URL", config: map[string]interface{}{"method": "GET", "url": "file:///etc/passwd"}, wantErr: true},
		{name: "URL credentials", config: map[string]interface{}{"method": "GET", "url": "https://user:pass@example.com"}, wantErr: true},
		{name: "unknown property", config: map[string]interface{}{"method": "GET", "url": "https://example.com", "template": "{{ secret }}"}, wantErr: true},
		{name: "forbidden header", config: map[string]interface{}{"method": "GET", "url": "https://example.com", "headers": map[string]interface{}{"Host": "internal"}}, wantErr: true},
		{name: "oversized body", config: map[string]interface{}{"method": "POST", "url": "https://example.com", "body": strings.Repeat("x", maxHTTPRequestBody+1)}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := api.Node{Id: "http-1", Type: HTTPNodeType, Config: &test.config}
			err := ValidateNodeConfig(node)
			if test.wantErr && err == nil {
				t.Fatal("ValidateNodeConfig() error = nil, want an error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateNodeConfig() error = %v", err)
			}
		})
	}
}

func TestHTTPURLPolicy(t *testing.T) {
	t.Parallel()

	policy, err := newHTTPURLPolicy([]string{"api.example.com"})
	if err != nil {
		t.Fatalf("newHTTPURLPolicy() error = %v", err)
	}
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{name: "allowlisted HTTPS", rawURL: "https://api.example.com/items"},
		{name: "unlisted hostname", rawURL: "https://other.example.com", wantErr: true},
		{name: "hostname suffix does not match", rawURL: "https://api.example.com.attacker.test", wantErr: true},
		{name: "userinfo rejected", rawURL: "https://api.example.com@attacker.test", wantErr: true},
		{name: "unsupported scheme", rawURL: "ftp://api.example.com/file", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request, requestErr := http.NewRequest(http.MethodGet, test.rawURL, nil)
			if requestErr != nil {
				if test.wantErr {
					return
				}
				t.Fatalf("http.NewRequest() error = %v", requestErr)
			}
			err := policy.validateURL(request.URL)
			if test.wantErr && err == nil {
				t.Fatal("validateURL() error = nil, want an error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("validateURL() error = %v", err)
			}
		})
	}
}

func TestHTTPURLPolicyRejectsPrivateDestinations(t *testing.T) {
	t.Parallel()

	for _, host := range []string{"localhost", "127.0.0.1", "10.0.0.1", "169.254.169.254", "::1"} {
		t.Run(host, func(t *testing.T) {
			t.Parallel()
			if _, err := newHTTPURLPolicy([]string{host}); err == nil {
				t.Fatalf("newHTTPURLPolicy(%q) error = nil, want an error", host)
			}
		})
	}
}

func TestHTTPURLPolicyBlocksPrivateAddressAtDialTime(t *testing.T) {
	t.Parallel()

	// Construct the policy directly to prove the dial boundary remains safe
	// even if a private address somehow bypasses allowlist configuration.
	policy := &httpURLPolicy{
		allowedHosts: map[string]struct{}{"127.0.0.1": {}},
		dialer:       &net.Dialer{},
	}
	if _, err := policy.dialContext(context.Background(), "tcp", "127.0.0.1:80"); err == nil {
		t.Fatal("dialContext() error = nil, want blocked-address error")
	}
}

func TestHTTPNodeActivityExecute(t *testing.T) {
	t.Parallel()

	policy, err := newHTTPURLPolicy([]string{"api.example.com"})
	if err != nil {
		t.Fatalf("newHTTPURLPolicy() error = %v", err)
	}
	activity := &HTTPNodeActivity{
		policy: policy,
		client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", request.Method)
			}
			if got := request.Header.Get("X-Request-ID"); got != "request-1" {
				t.Errorf("X-Request-ID = %q, want request-1", got)
			}
			body, readErr := io.ReadAll(request.Body)
			if readErr != nil {
				t.Fatalf("read request body: %v", readErr)
			}
			if got := string(body); got != `{"message":"hello"}` {
				t.Errorf("body = %q", got)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"item-1"}`)),
				Request:    request,
			}, nil
		})},
	}
	config := map[string]interface{}{
		"method":  "POST",
		"url":     "https://api.example.com/items",
		"headers": map[string]interface{}{"X-Request-ID": "request-1"},
		"body":    `{"message":"hello"}`,
	}
	result, err := activity.Execute(context.Background(), NodeActivityInput{
		Node: nodetype.Node{ID: "http-1", Type: HTTPNodeType, Config: &config},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.NodeID != "http-1" {
		t.Errorf("NodeID = %q, want http-1", result.NodeID)
	}
	if got := result.Value["statusCode"]; got != http.StatusCreated {
		t.Errorf("statusCode = %#v, want %d", got, http.StatusCreated)
	}
	if got := result.Value["body"]; got != `{"id":"item-1"}` {
		t.Errorf("body = %#v", got)
	}
}

func TestHTTPNodeActivityRejectsOversizedResponse(t *testing.T) {
	t.Parallel()

	policy, err := newHTTPURLPolicy([]string{"api.example.com"})
	if err != nil {
		t.Fatalf("newHTTPURLPolicy() error = %v", err)
	}
	activity := &HTTPNodeActivity{
		policy: policy,
		client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", maxHTTPResponseBody+1))),
				Request:    request,
			}, nil
		})},
	}
	config := map[string]interface{}{"method": "GET", "url": "https://api.example.com/large"}
	_, err = activity.Execute(context.Background(), NodeActivityInput{
		Node: nodetype.Node{ID: "http-1", Type: HTTPNodeType, Config: &config},
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want oversized response error")
	}
}

func TestHTTPNodeActivityDoesNotLeakRequestURLOnFailure(t *testing.T) {
	t.Parallel()

	policy, err := newHTTPURLPolicy([]string{"api.example.com"})
	if err != nil {
		t.Fatalf("newHTTPURLPolicy() error = %v", err)
	}
	activity := &HTTPNodeActivity{
		policy: policy,
		client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("GET https://api.example.com/items?token=super-secret failed")
		})},
	}
	config := map[string]interface{}{
		"method": "GET",
		"url":    "https://api.example.com/items?token=super-secret",
	}
	_, err = activity.Execute(context.Background(), NodeActivityInput{
		Node: nodetype.Node{ID: "http-1", Type: HTTPNodeType, Config: &config},
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want request failure")
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("Execute() error leaked URL query: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
