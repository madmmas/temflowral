package store

import (
	"context"
	"fmt"
	"sort"
	"sync"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// MemoryStore is an in-memory graph/run registry for tests. Process restart
// clears all state. Production must use a durable Store (see OpenFromEnv).
type MemoryStore struct {
	mu     sync.RWMutex
	graphs map[openapi_types.UUID]api.Graph
	runs   map[openapi_types.UUID]RunRecord
	// idempotency indexes (graphID, key) -> runID
	idempotency map[string]openapi_types.UUID
}

// NewMemoryStore returns an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		graphs:      make(map[openapi_types.UUID]api.Graph),
		runs:        make(map[openapi_types.UUID]RunRecord),
		idempotency: make(map[string]openapi_types.UUID),
	}
}

func (store *MemoryStore) PutGraph(_ context.Context, graph api.Graph) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.graphs[graph.Id] = graph
	return nil
}

func (store *MemoryStore) GetGraph(_ context.Context, id openapi_types.UUID) (api.Graph, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	graph, ok := store.graphs[id]
	return graph, ok, nil
}

func (store *MemoryStore) ListGraphs(_ context.Context) ([]api.GraphSummary, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	summaries := make([]api.GraphSummary, 0, len(store.graphs))
	for _, graph := range store.graphs {
		summaries = append(summaries, api.GraphSummary{
			Id:        graph.Id,
			Name:      graph.Name,
			CreatedAt: graph.CreatedAt,
			UpdatedAt: graph.UpdatedAt,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].UpdatedAt.Equal(summaries[j].UpdatedAt) {
			return summaries[i].Id.String() > summaries[j].Id.String()
		}
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})
	return summaries, nil
}

func (store *MemoryStore) UpdateGraph(_ context.Context, graph api.Graph) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.graphs[graph.Id]; !ok {
		return false, nil
	}
	store.graphs[graph.Id] = graph
	return true, nil
}

func (store *MemoryStore) DeleteGraph(_ context.Context, id openapi_types.UUID) (bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.graphs[id]; !ok {
		return false, nil
	}
	for runID, record := range store.runs {
		if record.Run.GraphId != id {
			continue
		}
		delete(store.runs, runID)
		if record.IdempotencyKey != nil {
			delete(store.idempotency, idempotencyIndexKey(id, *record.IdempotencyKey))
		}
	}
	delete(store.graphs, id)
	return true, nil
}

func (store *MemoryStore) PutRun(_ context.Context, record RunRecord) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if record.IdempotencyKey != nil {
		indexKey := idempotencyIndexKey(record.Run.GraphId, *record.IdempotencyKey)
		if existingID, ok := store.idempotency[indexKey]; ok && existingID != record.Run.Id {
			return fmt.Errorf("%w", ErrDuplicateIdempotencyKey)
		}
		store.idempotency[indexKey] = record.Run.Id
	}
	store.runs[record.Run.Id] = record
	return nil
}

func (store *MemoryStore) GetRun(_ context.Context, id openapi_types.UUID) (RunRecord, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, ok := store.runs[id]
	return record, ok, nil
}

func (store *MemoryStore) GetRunByIdempotencyKey(
	_ context.Context,
	graphID openapi_types.UUID,
	key string,
) (RunRecord, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	runID, ok := store.idempotency[idempotencyIndexKey(graphID, key)]
	if !ok {
		return RunRecord{}, false, nil
	}
	record, ok := store.runs[runID]
	return record, ok, nil
}

func (store *MemoryStore) UpdateRun(_ context.Context, record RunRecord) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.runs[record.Run.Id] = record
	if record.IdempotencyKey != nil {
		store.idempotency[idempotencyIndexKey(record.Run.GraphId, *record.IdempotencyKey)] = record.Run.Id
	}
	return nil
}

func (store *MemoryStore) Close() error {
	return nil
}

func idempotencyIndexKey(graphID openapi_types.UUID, key string) string {
	return graphID.String() + "\x00" + key
}
