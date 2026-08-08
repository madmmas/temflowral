package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madmmas/temflowral/backend/internal/store"
)

func TestParseCORSOrigins(t *testing.T) {
	t.Parallel()

	got := parseCORSOrigins(" http://localhost:3000, ,http://localhost:3000,http://127.0.0.1:3000 ")
	want := []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("origins[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAPICORSAllowsLocalFrontend(t *testing.T) {
	t.Parallel()

	handler := newHandler(
		[]byte("openapi: 3.1.0\n"),
		NewAPI(store.NewMemoryStore(), &stubRunner{}, nil),
		"",
		[]string{"http://localhost:3000"},
	)

	t.Run("preflight", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodOptions, "/node-types", nil)
		request.Header.Set("Origin", "http://localhost:3000")
		request.Header.Set("Access-Control-Request-Method", "GET")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Fatalf("Allow-Origin = %q, want localhost:3000", got)
		}
		if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
			t.Fatalf("Allow-Methods = %q, want GET", got)
		}
		if got := recorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
			t.Fatalf("Allow-Headers = %q, want Authorization", got)
		}
	})

	t.Run("get", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
		request.Header.Set("Origin", "http://localhost:3000")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Fatalf("Allow-Origin = %q, want localhost:3000", got)
		}
		if got := recorder.Header().Get("Vary"); !strings.Contains(got, "Origin") {
			t.Fatalf("Vary = %q, want Origin", got)
		}
	})

	t.Run("rejects unknown origin preflight", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodOptions, "/node-types", nil)
		request.Header.Set("Origin", "http://evil.example")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Allow-Origin = %q, want empty", got)
		}
	})

	t.Run("omits headers for unknown origin get", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
		request.Header.Set("Origin", "http://evil.example")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Allow-Origin = %q, want empty", got)
		}
	})
}

func TestAPICORSPreflightBypassesAuth(t *testing.T) {
	t.Parallel()

	handler := newHandler(
		[]byte("openapi: 3.1.0\n"),
		NewAPI(store.NewMemoryStore(), &stubRunner{}, nil),
		"test-secret",
		[]string{"http://localhost:3000"},
	)

	request := httptest.NewRequest(http.MethodOptions, "/node-types", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", "GET")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusNoContent, recorder.Body.String())
	}
}

func TestAPICORSOriginsFromEnv(t *testing.T) {
	t.Parallel()

	t.Run("defaults when unset", func(t *testing.T) {
		t.Parallel()
		got := corsOriginsFromEnv(func(string) (string, bool) { return "", false })
		if len(got) != len(defaultCORSOrigins) {
			t.Fatalf("origins = %v, want defaults %v", got, defaultCORSOrigins)
		}
		for i := range defaultCORSOrigins {
			if got[i] != defaultCORSOrigins[i] {
				t.Fatalf("origins[%d] = %q, want %q", i, got[i], defaultCORSOrigins[i])
			}
		}
	})

	t.Run("override list", func(t *testing.T) {
		t.Parallel()
		got := corsOriginsFromEnv(func(string) (string, bool) {
			return "https://app.example.com", true
		})
		if len(got) != 1 || got[0] != "https://app.example.com" {
			t.Fatalf("origins = %v, want [https://app.example.com]", got)
		}
	})

	t.Run("empty disables", func(t *testing.T) {
		t.Parallel()
		got := corsOriginsFromEnv(func(string) (string, bool) { return "", true })
		if len(got) != 0 {
			t.Fatalf("origins = %v, want []", got)
		}
	})
}
