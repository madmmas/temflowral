package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/madmmas/temflowral/backend/internal/server"
	temporalruntime "github.com/madmmas/temflowral/backend/internal/temporal"
)

const (
	listenAddress      = ":8080"
	openAPISpecPathEnv = "OPENAPI_SPEC_PATH"
	shutdownTimeout    = 10 * time.Second
)

var defaultSpecPaths = []string{
	"api/openapi.yaml",
	"../api/openapi.yaml",
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	openAPISpec, specPath, err := loadOpenAPISpec()
	if err != nil {
		return err
	}

	temporalConfig := temporalruntime.ConfigFromEnv()
	temporalRuntime, err := temporalruntime.Start(temporalConfig)
	if err != nil {
		return fmt.Errorf("initialize Temporal: %w", err)
	}
	defer temporalRuntime.Close()

	apiServer := server.NewAPI(server.NewStore(), temporalRuntime)
	httpServer := &http.Server{
		Addr:              listenAddress,
		Handler:           server.NewHandler(openAPISpec, apiServer),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf(
		"Temporal worker polling %s in namespace %s at %s",
		temporalConfig.TaskQueue,
		temporalConfig.Namespace,
		temporalConfig.Address,
	)
	log.Printf("serving API documentation at http://localhost%s/docs (contract: %s)", listenAddress, specPath)

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	select {
	case <-signalContext.Done():
		log.Print("shutting down backend")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	return nil
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
