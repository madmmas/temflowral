package server

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const apiAuthTokenEnv = "API_AUTH_TOKEN"

// apiAuthToken returns the configured shared secret, or empty when auth is off.
func apiAuthToken() string {
	return strings.TrimSpace(os.Getenv(apiAuthTokenEnv))
}

// bearerAuthMiddlewares returns OpenAPI handler middleware that requires
// Authorization: Bearer <token> when token is non-empty. Docs routes
// (/docs, /openapi.yaml) are registered outside this middleware.
func bearerAuthMiddlewares(token string) []api.MiddlewareFunc {
	if token == "" {
		return nil
	}
	return []api.MiddlewareFunc{requireBearerToken(token)}
}

func requireBearerToken(expected string) api.MiddlewareFunc {
	expectedBytes := []byte(expected)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok || !secureTokenEqual(got, expectedBytes) {
				writeUnauthorized(w, "missing or invalid bearer token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func secureTokenEqual(got string, expected []byte) bool {
	gotBytes := []byte(got)
	if len(gotBytes) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare(gotBytes, expected) == 1
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	code := "unauthorized"
	body := api.Error{Code: &code, Message: message}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="temflowral"`)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(body)
}
