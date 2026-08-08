// Command seed-demo upserts fixed-UUID demo graphs into the durable store.
//
//	DATABASE_URL=postgres://temporal:temporal@127.0.0.1:5432/temflowral?sslmode=disable \
//	  go run ./cmd/seed-demo
//
// Or from the repo root: make seed-demo (requires compose Postgres + a prior
// backend boot so the application schema exists).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/madmmas/temflowral/backend/internal/seed"
	"github.com/madmmas/temflowral/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	graphStore, err := store.OpenFromEnv()
	if err != nil {
		return err
	}
	defer func() { _ = graphStore.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	graphs, err := seed.UpsertDemoGraphs(ctx, graphStore, time.Now().UTC())
	if err != nil {
		return err
	}

	base := os.Getenv("FRONTEND_BASE_URL")
	if base == "" {
		base = "http://localhost:3000"
	}

	fmt.Println("Demo graphs upserted:")
	for _, graph := range graphs {
		name := "Untitled"
		if graph.Name != nil && *graph.Name != "" {
			name = *graph.Name
		}
		fmt.Printf("  %s\n    id:  %s\n    open: %s/?graph=%s\n", name, graph.Id, base, graph.Id)
	}
	return nil
}
