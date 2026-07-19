package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler(t *testing.T) {
	t.Parallel()

	const spec = "openapi: 3.1.0\ninfo:\n  title: temflowral API\n"
	handler := NewDocsHandler([]byte(spec))

	tests := []struct {
		name        string
		path        string
		contentType string
		body        string
	}{
		{
			name:        "serves OpenAPI contract",
			path:        "/openapi.yaml",
			contentType: "application/yaml; charset=utf-8",
			body:        spec,
		},
		{
			name:        "serves Swagger UI",
			path:        "/docs",
			contentType: "text/html; charset=utf-8",
			body:        `url: "/openapi.yaml"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			if got := response.Header.Get("Content-Type"); got != test.contentType {
				t.Errorf("Content-Type = %q, want %q", got, test.contentType)
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if !strings.Contains(string(body), test.body) {
				t.Errorf("response body does not contain %q", test.body)
			}
		})
	}
}

func TestDocsHandlerRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/docs", nil)
	recorder := httptest.NewRecorder()
	NewDocsHandler(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
