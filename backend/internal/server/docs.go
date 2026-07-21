package server

import (
	"bytes"
	"net/http"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>temflowral API documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true
      });
    };
  </script>
</body>
</html>
`

// NewHandler serves the generated API routes, OpenAPI contract, and Swagger UI.
// When API_AUTH_TOKEN is set, contract routes require Authorization: Bearer;
// /docs and /openapi.yaml stay reachable without a token.
func NewHandler(openAPISpec []byte, implementation api.StrictServerInterface) http.Handler {
	return newHandler(openAPISpec, implementation, apiAuthToken())
}

func newHandler(openAPISpec []byte, implementation api.StrictServerInterface, authToken string) http.Handler {
	mux := http.NewServeMux()
	registerDocsRoutes(mux, openAPISpec)

	strictHandler := api.NewStrictHandler(implementation, nil)
	return api.HandlerWithOptions(strictHandler, api.StdHTTPServerOptions{
		BaseRouter:  mux,
		Middlewares: bearerAuthMiddlewares(authToken),
	})
}

// NewDocsHandler serves only the OpenAPI contract and its Swagger UI.
func NewDocsHandler(openAPISpec []byte) http.Handler {
	mux := http.NewServeMux()
	registerDocsRoutes(mux, openAPISpec)
	return mux
}

func registerDocsRoutes(mux *http.ServeMux, openAPISpec []byte) {
	spec := bytes.Clone(openAPISpec)

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		if _, err := w.Write(spec); err != nil {
			return
		}
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(swaggerUI)); err != nil {
			return
		}
	})
}
