package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/madmmas/temflowral/backend/internal/seed"
	"github.com/madmmas/temflowral/backend/internal/store"
	"github.com/madmmas/temflowral/backend/internal/temporal"
)

func TestUpsertDemoGraphsIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	memory := store.NewMemoryStore()
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)

	first, err := seed.UpsertDemoGraphs(ctx, memory, now)
	if err != nil {
		t.Fatalf("UpsertDemoGraphs() error = %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("len = %d, want 2", len(first))
	}

	second, err := seed.UpsertDemoGraphs(ctx, memory, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("second UpsertDemoGraphs() error = %v", err)
	}
	if second[0].Id != seed.DemoStartNoopID || second[1].Id != seed.DemoStartDelayNoopID {
		t.Fatalf("ids = %v %v", second[0].Id, second[1].Id)
	}

	got, ok, err := memory.GetGraph(ctx, seed.DemoStartNoopID)
	if err != nil || !ok {
		t.Fatalf("GetGraph start-noop ok=%v err=%v", ok, err)
	}
	if err := temporal.ValidateGraph(got, nil); err != nil {
		t.Fatalf("ValidateGraph start-noop: %v", err)
	}

	gotDelay, ok, err := memory.GetGraph(ctx, seed.DemoStartDelayNoopID)
	if err != nil || !ok {
		t.Fatalf("GetGraph start-delay-noop ok=%v err=%v", ok, err)
	}
	if err := temporal.ValidateGraph(gotDelay, nil); err != nil {
		t.Fatalf("ValidateGraph start-delay-noop: %v", err)
	}
}
