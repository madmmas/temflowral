package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOpenAPISpecFromEnvironment(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "openapi.yaml")
	const spec = "openapi: 3.1.0\n"
	if err := os.WriteFile(specPath, []byte(spec), 0o600); err != nil {
		t.Fatalf("write OpenAPI fixture: %v", err)
	}
	t.Setenv(openAPISpecPathEnv, specPath)

	gotSpec, gotPath, err := loadOpenAPISpec()
	if err != nil {
		t.Fatalf("loadOpenAPISpec() error = %v", err)
	}
	if string(gotSpec) != spec {
		t.Errorf("spec = %q, want %q", gotSpec, spec)
	}
	if gotPath != specPath {
		t.Errorf("path = %q, want %q", gotPath, specPath)
	}
}

func TestLoadOpenAPISpecReportsInvalidEnvironmentPath(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "missing.yaml")
	t.Setenv(openAPISpecPathEnv, specPath)

	_, _, err := loadOpenAPISpec()
	if err == nil {
		t.Fatal("loadOpenAPISpec() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), specPath) {
		t.Errorf("error %q does not contain path %q", err, specPath)
	}
}
