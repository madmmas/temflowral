package temporal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// HTTPNodeType executes one allowlisted outbound HTTP request.
	HTTPNodeType = "http"
	// HTTPNodeActivityName is the Temporal activity type for HTTP nodes.
	HTTPNodeActivityName = "temflowral.node.http"

	maxHTTPURLLength      = 2048
	maxHTTPRequestBody    = 1 << 20 // 1 MiB
	maxHTTPResponseBody   = 1 << 20 // 1 MiB
	maxHTTPHeaders        = 32
	maxHTTPHeaderValue    = 8192
	httpActivityTimeout   = 20 * time.Second
	maxResponseHeaderSize = 64 << 10 // 64 KiB
)

var (
	allowedHTTPMethods = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}
	forbiddenHTTPHeaders = map[string]struct{}{
		"Connection":          {},
		"Content-Length":      {},
		"Host":                {},
		"Proxy-Authorization": {},
		"Proxy-Connection":    {},
		"Te":                  {},
		"Trailer":             {},
		"Transfer-Encoding":   {},
		"Upgrade":             {},
	}
)

// HTTPNodeActivity executes HTTP nodes with an injected client and URL policy.
type HTTPNodeActivity struct {
	client *http.Client
	policy *httpURLPolicy
}

// NewHTTPNodeActivity creates a hardened HTTP activity. allowedHosts must
// contain exact hostnames or public IP literals; an empty list denies all.
func NewHTTPNodeActivity(allowedHosts []string) (*HTTPNodeActivity, error) {
	policy, err := newHTTPURLPolicy(allowedHosts)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = policy.dialContext
	transport.MaxResponseHeaderBytes = maxResponseHeaderSize

	return &HTTPNodeActivity{
		policy: policy,
		client: &http.Client{
			Transport: transport,
			Timeout:   httpActivityTimeout,
			CheckRedirect: func(request *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("too many redirects")
				}
				return policy.validateURL(request.URL)
			},
		},
	}, nil
}

// Execute performs the configured request and returns a bounded response.
func (activity *HTTPNodeActivity) Execute(ctx context.Context, input NodeActivityInput) (NodeResult, error) {
	config, err := parseHTTPNodeConfig(input.Node)
	if err != nil {
		return NodeResult{}, err
	}
	requestURL, err := url.Parse(config.Url)
	if err != nil {
		return NodeResult{}, errors.New("parse HTTP node URL")
	}
	if err := activity.policy.validateURL(requestURL); err != nil {
		return NodeResult{}, err
	}

	body := ""
	if config.Body != nil {
		body = *config.Body
	}
	request, err := http.NewRequestWithContext(
		ctx,
		string(config.Method),
		config.Url,
		bytes.NewBufferString(body),
	)
	if err != nil {
		return NodeResult{}, errors.New("create HTTP node request")
	}
	if config.Headers != nil {
		for name, value := range *config.Headers {
			request.Header.Set(name, value)
		}
	}

	response, err := activity.client.Do(request)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return NodeResult{}, errors.New("HTTP node request cancelled or timed out")
		}
		// net/http errors can contain the complete URL (including sensitive
		// query parameters), so do not propagate the underlying message.
		return NodeResult{}, errors.New("HTTP node request failed")
	}

	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, maxHTTPResponseBody+1))
	closeErr := response.Body.Close()
	if readErr != nil {
		return NodeResult{}, fmt.Errorf("read HTTP node response: %w", readErr)
	}
	if closeErr != nil {
		return NodeResult{}, fmt.Errorf("close HTTP node response: %w", closeErr)
	}
	if len(responseBody) > maxHTTPResponseBody {
		return NodeResult{}, fmt.Errorf("HTTP node response exceeds %d bytes", maxHTTPResponseBody)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return NodeResult{}, fmt.Errorf("HTTP node request returned status %d", response.StatusCode)
	}

	value := map[string]interface{}{
		"statusCode": response.StatusCode,
		"body":       string(responseBody),
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		value["contentType"] = contentType
	}
	return NodeResult{NodeID: input.Node.Id, Value: value}, nil
}

// ValidateNodeConfig validates node-type-specific configuration without
// performing deployment-specific allowlist or DNS checks.
func ValidateNodeConfig(node api.Node) error {
	switch node.Type {
	case HTTPNodeType:
		_, err := parseHTTPNodeConfig(node)
		return err
	case DelayNodeType:
		_, err := parseDelayNodeConfig(node)
		return err
	case ConditionNodeType:
		_, err := parseConditionNodeConfig(node)
		return err
	default:
		return nil
	}
}

func parseHTTPNodeConfig(node api.Node) (api.HttpNodeConfig, error) {
	if node.Config == nil {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q config is required", node.Id)
	}
	encoded, err := json.Marshal(*node.Config)
	if err != nil {
		return api.HttpNodeConfig{}, fmt.Errorf("encode HTTP node %q config: %w", node.Id, err)
	}

	var config api.HttpNodeConfig
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return api.HttpNodeConfig{}, fmt.Errorf("invalid HTTP node %q config: %w", node.Id, err)
	}
	if !config.Method.Valid() || !slices.Contains(allowedHTTPMethods, string(config.Method)) {
		return api.HttpNodeConfig{}, fmt.Errorf(
			"HTTP node %q method must be one of %s",
			node.Id,
			strings.Join(allowedHTTPMethods, ", "),
		)
	}
	if len(config.Url) == 0 || len(config.Url) > maxHTTPURLLength {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q URL must be 1-%d bytes", node.Id, maxHTTPURLLength)
	}
	parsedURL, err := url.Parse(config.Url)
	if err != nil || !parsedURL.IsAbs() {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q URL must be absolute", node.Id)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q URL scheme must be http or https", node.Id)
	}
	if parsedURL.Hostname() == "" {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q URL must include a hostname", node.Id)
	}
	if parsedURL.User != nil {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q URL must not contain user information", node.Id)
	}
	if config.Body != nil && len(*config.Body) > maxHTTPRequestBody {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q body exceeds %d bytes", node.Id, maxHTTPRequestBody)
	}
	if config.Headers != nil && len(*config.Headers) > maxHTTPHeaders {
		return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q has more than %d headers", node.Id, maxHTTPHeaders)
	}
	if config.Headers != nil {
		for name, value := range *config.Headers {
			canonicalName := http.CanonicalHeaderKey(name)
			if !httpguts.ValidHeaderFieldName(name) || !httpguts.ValidHeaderFieldValue(value) {
				return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q contains an invalid header", node.Id)
			}
			if _, forbidden := forbiddenHTTPHeaders[canonicalName]; forbidden {
				return api.HttpNodeConfig{}, fmt.Errorf("HTTP node %q header %q is not allowed", node.Id, canonicalName)
			}
			if len(value) > maxHTTPHeaderValue {
				return api.HttpNodeConfig{}, fmt.Errorf(
					"HTTP node %q header %q exceeds %d bytes",
					node.Id,
					canonicalName,
					maxHTTPHeaderValue,
				)
			}
		}
	}
	return config, nil
}

type httpURLPolicy struct {
	allowedHosts map[string]struct{}
	resolver     *net.Resolver
	dialer       *net.Dialer
}

func newHTTPURLPolicy(allowedHosts []string) (*httpURLPolicy, error) {
	policy := &httpURLPolicy{
		allowedHosts: make(map[string]struct{}, len(allowedHosts)),
		resolver:     net.DefaultResolver,
		dialer:       &net.Dialer{Timeout: 10 * time.Second},
	}
	for _, allowedHost := range allowedHosts {
		host := canonicalHostname(allowedHost)
		if host == "" {
			return nil, fmt.Errorf("invalid HTTP allowed host %q: use a hostname without scheme or port", allowedHost)
		}
		if host == "localhost" || strings.HasSuffix(host, ".localhost") {
			return nil, fmt.Errorf("HTTP allowed host %q is local and cannot be permitted", allowedHost)
		}
		if ip := net.ParseIP(host); ip != nil {
			if isBlockedDestination(ip) {
				return nil, fmt.Errorf("HTTP allowed host %q is not a public address", allowedHost)
			}
		} else if strings.ContainsAny(host, "/:@") {
			return nil, fmt.Errorf("invalid HTTP allowed host %q: use a hostname without scheme or port", allowedHost)
		}
		policy.allowedHosts[host] = struct{}{}
	}
	return policy, nil
}

func (policy *httpURLPolicy) validateURL(requestURL *url.URL) error {
	if requestURL == nil || !requestURL.IsAbs() || requestURL.Hostname() == "" {
		return errors.New("HTTP node URL must be absolute")
	}
	if len(requestURL.String()) > maxHTTPURLLength {
		return fmt.Errorf("HTTP node URL exceeds %d bytes", maxHTTPURLLength)
	}
	if requestURL.Scheme != "http" && requestURL.Scheme != "https" {
		return errors.New("HTTP node URL scheme must be http or https")
	}
	if requestURL.User != nil {
		return errors.New("HTTP node URL must not contain user information")
	}
	host := canonicalHostname(requestURL.Hostname())
	if _, allowed := policy.allowedHosts[host]; !allowed {
		return fmt.Errorf("HTTP node host %q is not allowlisted", host)
	}
	if ip := net.ParseIP(host); ip != nil && isBlockedDestination(ip) {
		return fmt.Errorf("HTTP node host %q resolves to a blocked address", host)
	}
	return nil
}

func (policy *httpURLPolicy) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("parse HTTP destination %q: %w", address, err)
	}
	host = canonicalHostname(host)
	if _, allowed := policy.allowedHosts[host]; !allowed {
		return nil, fmt.Errorf("HTTP destination host %q is not allowlisted", host)
	}

	var addresses []net.IP
	if literal := net.ParseIP(host); literal != nil {
		addresses = []net.IP{literal}
	} else {
		resolved, resolveErr := policy.resolver.LookupIPAddr(ctx, host)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolve HTTP destination %q: %w", host, resolveErr)
		}
		for _, candidate := range resolved {
			addresses = append(addresses, candidate.IP)
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("HTTP destination %q resolved to no addresses", host)
	}
	for _, addressIP := range addresses {
		if isBlockedDestination(addressIP) {
			return nil, fmt.Errorf("HTTP destination %q resolves to a blocked address", host)
		}
	}

	var dialErrors []error
	for _, addressIP := range addresses {
		connection, dialErr := policy.dialer.DialContext(
			ctx,
			network,
			net.JoinHostPort(addressIP.String(), port),
		)
		if dialErr == nil {
			return connection, nil
		}
		dialErrors = append(dialErrors, dialErr)
	}
	return nil, fmt.Errorf("dial HTTP destination %q: %w", host, errors.Join(dialErrors...))
}

func canonicalHostname(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func isBlockedDestination(ip net.IP) bool {
	return !ip.IsGlobalUnicast() ||
		ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
