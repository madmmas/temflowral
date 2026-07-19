package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratedAPIRoutesUseStrictImplementation(t *testing.T) {
	t.Parallel()

	const id = "550e8400-e29b-41d4-a716-446655440000"
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI())

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create graph", method: http.MethodPost, path: "/graphs", body: `{}`},
		{name: "get graph", method: http.MethodGet, path: "/graphs/" + id},
		{name: "start graph run", method: http.MethodPost, path: "/graphs/" + id + "/run", body: `{}`},
		{name: "list node types", method: http.MethodGet, path: "/node-types"},
		{name: "get run", method: http.MethodGet, path: "/runs/" + id},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer response.Body.Close()

			if response.StatusCode != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
			}
			if got := response.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if !strings.Contains(string(body), `"code":"not_implemented"`) {
				t.Errorf("response body = %q, want not_implemented error", body)
			}
		})
	}
}
