package server

import (
	"sync"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// Store is an in-memory graph/run registry for early development. Process
// restart clears all state.
type Store struct {
	mu     sync.RWMutex
	graphs map[openapi_types.UUID]api.Graph
	runs   map[openapi_types.UUID]RunRecord
}

// RunRecord keeps the public Run payload plus Temporal execution IDs.
type RunRecord struct {
	Run                api.Run
	TemporalWorkflowID string
	TemporalRunID      string
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		graphs: make(map[openapi_types.UUID]api.Graph),
		runs:   make(map[openapi_types.UUID]RunRecord),
	}
}

func (store *Store) PutGraph(graph api.Graph) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.graphs[graph.Id] = graph
}

func (store *Store) GetGraph(id openapi_types.UUID) (api.Graph, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	graph, ok := store.graphs[id]
	return graph, ok
}

func (store *Store) PutRun(record RunRecord) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.runs[record.Run.Id] = record
}

func (store *Store) GetRun(id openapi_types.UUID) (RunRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, ok := store.runs[id]
	return record, ok
}

func (store *Store) UpdateRun(record RunRecord) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.runs[record.Run.Id] = record
}

func newGraphID() openapi_types.UUID {
	return uuid.New()
}

func newRunID() openapi_types.UUID {
	return uuid.New()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
