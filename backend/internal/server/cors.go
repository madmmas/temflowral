package server

import (
	"net/http"
	"os"
	"strings"
)

const apiCORSOriginsEnv = "API_CORS_ORIGINS"

// defaultCORSOrigins allow the reference canvas during local compose / make run.
// Override with API_CORS_ORIGINS in shared or production deployments.
var defaultCORSOrigins = []string{
	"http://localhost:3000",
	"http://127.0.0.1:3000",
}

// apiCORSOrigins returns the allowed browser Origin values.
// When API_CORS_ORIGINS is unset, defaults to the local frontend origins.
// When set (including empty), the comma-separated list is used as-is so
// operators can disable CORS with API_CORS_ORIGINS=.
func apiCORSOrigins() []string {
	return corsOriginsFromEnv(os.LookupEnv)
}

func corsOriginsFromEnv(lookup func(string) (string, bool)) []string {
	raw, ok := lookup(apiCORSOriginsEnv)
	if !ok {
		return append([]string(nil), defaultCORSOrigins...)
	}
	return parseCORSOrigins(raw)
}

func parseCORSOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins
}

// withCORS wraps next so browsers on allowed origins can call the API.
// Handles OPTIONS preflight outside auth middleware so preflight never
// requires a Bearer token.
func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		if r.Method == http.MethodOptions {
			if _, ok := allowed[origin]; ok {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
