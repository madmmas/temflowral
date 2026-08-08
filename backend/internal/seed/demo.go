// Package seed installs local demo graphs with fixed IDs for compose smoke (#94).
package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/internal/store"
)

// Fixed local-only demo graph IDs. Stable across `make seed-demo` runs so
// README / ?graph= links stay valid.
var (
	DemoStartNoopID      = uuid.MustParse("11111111-1111-4111-8111-111111111101")
	DemoStartDelayNoopID = uuid.MustParse("11111111-1111-4111-8111-111111111102")
)

// DemoGraphs returns the Start→No-op and Start→Delay→No-op fixtures.
func DemoGraphs(now time.Time) []api.Graph {
	now = now.UTC()
	startNoopName := "Demo: Start → No-op"
	startDelayName := "Demo: Start → Delay → No-op"
	seconds := 1.0

	return []api.Graph{
		{
			Id:   DemoStartNoopID,
			Name: &startNoopName,
			Nodes: []api.Node{
				{Id: "start-1", Type: "start", Label: strPtr("Start"), Position: api.Position{X: 0, Y: 0}, Config: &map[string]interface{}{}},
				{Id: "noop-1", Type: "noop", Label: strPtr("No-op"), Position: api.Position{X: 220, Y: 0}, Config: &map[string]interface{}{}},
			},
			Edges: []api.Edge{
				{Id: "e-start-noop", Source: "start-1", Target: "noop-1"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Id:   DemoStartDelayNoopID,
			Name: &startDelayName,
			Nodes: []api.Node{
				{Id: "start-1", Type: "start", Label: strPtr("Start"), Position: api.Position{X: 0, Y: 0}, Config: &map[string]interface{}{}},
				{
					Id:       "delay-1",
					Type:     "delay",
					Label:    strPtr("Delay 1s"),
					Position: api.Position{X: 220, Y: 0},
					Config:   &map[string]interface{}{"seconds": seconds},
				},
				{Id: "noop-1", Type: "noop", Label: strPtr("No-op"), Position: api.Position{X: 440, Y: 0}, Config: &map[string]interface{}{}},
			},
			Edges: []api.Edge{
				{Id: "e-start-delay", Source: "start-1", Target: "delay-1"},
				{Id: "e-delay-noop", Source: "delay-1", Target: "noop-1"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

// UpsertDemoGraphs writes demo graphs idempotently (PutGraph overwrite).
func UpsertDemoGraphs(ctx context.Context, graphStore store.Store, now time.Time) ([]api.Graph, error) {
	graphs := DemoGraphs(now)
	for _, graph := range graphs {
		if err := graphStore.PutGraph(ctx, graph); err != nil {
			return nil, fmt.Errorf("put demo graph %s: %w", graph.Id, err)
		}
	}
	return graphs, nil
}

func strPtr(value string) *string {
	return &value
}
