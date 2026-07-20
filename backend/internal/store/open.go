package store

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"
)

//go:embed schema.sql
var schemaSQL string

const (
	databaseURLEnv     = "DATABASE_URL"
	allowMemoryEnv     = "STORE_ALLOW_MEMORY"
	openTimeout        = 15 * time.Second
	errMissingDatabase = "DATABASE_URL is required: set a Postgres connection string " +
		"(e.g. postgres://user:pass@host:5432/temflowral). " +
		"In-memory storage is not used in production; set STORE_ALLOW_MEMORY=1 only for local experiments"
)

// OpenFromEnv opens the configured durable store. DATABASE_URL is required
// unless STORE_ALLOW_MEMORY=1 (tests / explicit local experiments only).
func OpenFromEnv() (Store, error) {
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnv))
	if databaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
		defer cancel()
		return OpenPostgres(ctx, databaseURL)
	}
	if allowMemory() {
		return NewMemoryStore(), nil
	}
	return nil, fmt.Errorf("%s", errMissingDatabase)
}

func allowMemory() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(allowMemoryEnv)))
	return value == "1" || value == "true" || value == "yes"
}
