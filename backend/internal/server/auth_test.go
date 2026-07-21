package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madmmas/temflowral/backend/internal/store"
)

func TestAPIAuthTokenRequiredWhenConfigured(t *testing.T) {
	t.Parallel()

	handler := newHandler(
		[]byte("openapi: 3.1.0\n"),
		NewAPI(store.NewMemoryStore(), &stubRunner{}, nil),
		"test-secret",
	)

	t.Run("rejects missing bearer", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
		if got := recorder.Header().Get("WWW-Authenticate"); !strings.Contains(got, "Bearer") {
			t.Fatalf("WWW-Authenticate = %q, want Bearer challenge", got)
		}
		var body map[string]interface{}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["code"] != "unauthorized" {
			t.Fatalf("code = %v, want unauthorized", body["code"])
		}
	})

	t.Run("rejects wrong bearer", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
		request.Header.Set("Authorization", "Bearer wrong")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
		}
	})

	t.Run("accepts matching bearer", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
		request.Header.Set("Authorization", "Bearer test-secret")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
		}
	})

	t.Run("docs remain public", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest(http.MethodGet, "/docs", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}
	})
}

func TestAPIAuthOpenWhenTokenUnset(t *testing.T) {
	t.Parallel()

	handler := newHandler(
		[]byte("openapi: 3.1.0\n"),
		NewAPI(store.NewMemoryStore(), &stubRunner{}, nil),
		"",
	)
	request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}
