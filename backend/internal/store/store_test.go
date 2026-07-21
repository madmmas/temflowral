package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestOpenFromEnvRequiresDatabaseURL(t *testing.T) {
	t.Setenv(databaseURLEnv, "")
	t.Setenv(allowMemoryEnv, "")
	_, err := OpenFromEnv()
	if err == nil {
		t.Fatal("OpenFromEnv() error = nil, want missing DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("OpenFromEnv() error = %v, want DATABASE_URL required", err)
	}
}

func TestOpenFromEnvAllowsExplicitMemory(t *testing.T) {
	t.Setenv(databaseURLEnv, "")
	t.Setenv(allowMemoryEnv, "1")
	opened, err := OpenFromEnv()
	if err != nil {
		t.Fatalf("OpenFromEnv() error = %v", err)
	}
	defer opened.Close()
	if _, ok := opened.(*MemoryStore); !ok {
		t.Fatalf("OpenFromEnv() type = %T, want *MemoryStore", opened)
	}
}

func TestMemoryStoreRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	memory := NewMemoryStore()
	name := "demo"
	graphID := uuid.New()
	graph := api.Graph{
		Id:        graphID,
		Name:      &name,
		Nodes:     []api.Node{{Id: "start-1", Type: "start"}},
		Edges:     []api.Edge{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := memory.PutGraph(ctx, graph); err != nil {
		t.Fatalf("PutGraph() error = %v", err)
	}
	got, ok, err := memory.GetGraph(ctx, graphID)
	if err != nil || !ok {
		t.Fatalf("GetGraph() ok=%v err=%v", ok, err)
	}
	if got.Id != graphID || got.Name == nil || *got.Name != name {
		t.Fatalf("GetGraph() = %#v", got)
	}

	runID := uuid.New()
	record := RunRecord{
		Run: api.Run{
			Id:        runID,
			GraphId:   graphID,
			Status:    api.Running,
			StartedAt: time.Now().UTC(),
		},
		TemporalWorkflowID: runID.String(),
		TemporalRunID:      "temporal-run-1",
	}
	if err := memory.PutRun(ctx, record); err != nil {
		t.Fatalf("PutRun() error = %v", err)
	}
	record.Run.Status = api.Completed
	if err := memory.UpdateRun(ctx, record); err != nil {
		t.Fatalf("UpdateRun() error = %v", err)
	}
	gotRun, ok, err := memory.GetRun(ctx, runID)
	if err != nil || !ok {
		t.Fatalf("GetRun() ok=%v err=%v", ok, err)
	}
	if gotRun.Run.Status != api.Completed || gotRun.TemporalRunID != "temporal-run-1" {
		t.Fatalf("GetRun() = %#v", gotRun)
	}
}

func TestMemoryStoreIdempotencyKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	memory := NewMemoryStore()
	graphID := uuid.New()
	if err := memory.PutGraph(ctx, api.Graph{
		Id:        graphID,
		Nodes:     []api.Node{},
		Edges:     []api.Edge{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("PutGraph() error = %v", err)
	}

	key := "delivery-1"
	runID := uuid.New()
	if err := memory.PutRun(ctx, RunRecord{
		Run: api.Run{
			Id:        runID,
			GraphId:   graphID,
			Status:    api.Running,
			StartedAt: time.Now().UTC(),
		},
		TemporalWorkflowID: runID.String(),
		TemporalRunID:      "t1",
		IdempotencyKey:     &key,
	}); err != nil {
		t.Fatalf("PutRun() error = %v", err)
	}

	got, ok, err := memory.GetRunByIdempotencyKey(ctx, graphID, key)
	if err != nil || !ok || got.Run.Id != runID {
		t.Fatalf("GetRunByIdempotencyKey() = %#v ok=%v err=%v", got, ok, err)
	}

	otherID := uuid.New()
	err = memory.PutRun(ctx, RunRecord{
		Run: api.Run{
			Id:        otherID,
			GraphId:   graphID,
			Status:    api.Running,
			StartedAt: time.Now().UTC(),
		},
		TemporalWorkflowID: otherID.String(),
		TemporalRunID:      "t2",
		IdempotencyKey:     &key,
	})
	if !errors.Is(err, ErrDuplicateIdempotencyKey) {
		t.Fatalf("PutRun() duplicate error = %v, want ErrDuplicateIdempotencyKey", err)
	}
}

func TestPostgresStoreRoundTrip(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnv))
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pgStore, err := OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatalf("OpenPostgres() error = %v", err)
	}
	defer pgStore.Close()

	name := "postgres-demo"
	graphID := uuid.New()
	graph := api.Graph{
		Id:   graphID,
		Name: &name,
		Nodes: []api.Node{
			{Id: "start-1", Type: "start", Position: api.Position{X: 0, Y: 0}},
			{Id: "noop-1", Type: "noop", Position: api.Position{X: 1, Y: 0}},
		},
		Edges:     []api.Edge{{Id: "e1", Source: "start-1", Target: "noop-1"}},
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := pgStore.PutGraph(ctx, graph); err != nil {
		t.Fatalf("PutGraph() error = %v", err)
	}
	got, ok, err := pgStore.GetGraph(ctx, graphID)
	if err != nil || !ok {
		t.Fatalf("GetGraph() ok=%v err=%v", ok, err)
	}
	if got.Id != graphID || len(got.Nodes) != 2 || len(got.Edges) != 1 {
		t.Fatalf("GetGraph() = %#v", got)
	}

	runID := uuid.New()
	result := map[string]interface{}{"nodes": []interface{}{}}
	record := RunRecord{
		Run: api.Run{
			Id:        runID,
			GraphId:   graphID,
			Status:    api.Completed,
			StartedAt: time.Now().UTC().Truncate(time.Microsecond),
			Result:    &result,
		},
		TemporalWorkflowID: runID.String(),
		TemporalRunID:      "pg-run-1",
	}
	completed := time.Now().UTC().Truncate(time.Microsecond)
	record.Run.CompletedAt = &completed
	if err := pgStore.PutRun(ctx, record); err != nil {
		t.Fatalf("PutRun() error = %v", err)
	}
	gotRun, ok, err := pgStore.GetRun(ctx, runID)
	if err != nil || !ok {
		t.Fatalf("GetRun() ok=%v err=%v", ok, err)
	}
	if gotRun.Run.Status != api.Completed || gotRun.TemporalRunID != "pg-run-1" || gotRun.Run.Result == nil {
		t.Fatalf("GetRun() = %#v", gotRun)
	}
}

func TestPostgresStoreIdempotencyKey(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv(databaseURLEnv))
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pgStore, err := OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatalf("OpenPostgres() error = %v", err)
	}
	defer pgStore.Close()

	graphID := uuid.New()
	if err := pgStore.PutGraph(ctx, api.Graph{
		Id:        graphID,
		Nodes:     []api.Node{},
		Edges:     []api.Edge{},
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}); err != nil {
		t.Fatalf("PutGraph() error = %v", err)
	}

	key := "pg-delivery-" + uuid.NewString()
	runID := uuid.New()
	if err := pgStore.PutRun(ctx, RunRecord{
		Run: api.Run{
			Id:        runID,
			GraphId:   graphID,
			Status:    api.Running,
			StartedAt: time.Now().UTC().Truncate(time.Microsecond),
		},
		TemporalWorkflowID: runID.String(),
		TemporalRunID:      "pg-t1",
		IdempotencyKey:     &key,
	}); err != nil {
		t.Fatalf("PutRun() error = %v", err)
	}

	got, ok, err := pgStore.GetRunByIdempotencyKey(ctx, graphID, key)
	if err != nil || !ok || got.Run.Id != runID {
		t.Fatalf("GetRunByIdempotencyKey() = %#v ok=%v err=%v", got, ok, err)
	}

	otherID := uuid.New()
	err = pgStore.PutRun(ctx, RunRecord{
		Run: api.Run{
			Id:        otherID,
			GraphId:   graphID,
			Status:    api.Running,
			StartedAt: time.Now().UTC().Truncate(time.Microsecond),
		},
		TemporalWorkflowID: otherID.String(),
		TemporalRunID:      "pg-t2",
		IdempotencyKey:     &key,
	})
	if !errors.Is(err, ErrDuplicateIdempotencyKey) {
		t.Fatalf("PutRun() duplicate error = %v, want ErrDuplicateIdempotencyKey", err)
	}
}
