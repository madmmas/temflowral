package store

import (
	"context"
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
}

// NewMemoryStore returns an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		graphs: make(map[openapi_types.UUID]api.Graph),
		runs:   make(map[openapi_types.UUID]RunRecord),
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

func (store *MemoryStore) PutRun(_ context.Context, record RunRecord) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.runs[record.Run.Id] = record
	return nil
}

func (store *MemoryStore) GetRun(_ context.Context, id openapi_types.UUID) (RunRecord, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, ok := store.runs[id]
	return record, ok, nil
}

func (store *MemoryStore) UpdateRun(_ context.Context, record RunRecord) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.runs[record.Run.Id] = record
	return nil
}

func (store *MemoryStore) Close() error {
	return nil
}
