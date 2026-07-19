package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/madmmas/temflowral/backend/internal/server"
)

const (
	listenAddress      = ":8080"
	openAPISpecPathEnv = "OPENAPI_SPEC_PATH"
)

var defaultSpecPaths = []string{
	"api/openapi.yaml",
	"../api/openapi.yaml",
}

func main() {
	openAPISpec, specPath, err := loadOpenAPISpec()
	if err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:              listenAddress,
		Handler:           server.NewDocsHandler(openAPISpec),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("serving API documentation at http://localhost%s/docs (contract: %s)", listenAddress, specPath)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadOpenAPISpec() ([]byte, string, error) {
	if specPath := os.Getenv(openAPISpecPathEnv); specPath != "" {
		spec, err := os.ReadFile(specPath)
		if err != nil {
			return nil, "", fmt.Errorf("read OpenAPI contract from %s: %w", specPath, err)
		}
		return spec, specPath, nil
	}

	for _, specPath := range defaultSpecPaths {
		spec, err := os.ReadFile(specPath)
		if err == nil {
			return spec, specPath, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("read OpenAPI contract from %s: %w", specPath, err)
		}
	}

	return nil, "", fmt.Errorf(
		"OpenAPI contract not found; set %s or run the server from the repository root or backend directory",
		openAPISpecPathEnv,
	)
}
