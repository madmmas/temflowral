package store

import (
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// RunRecord keeps the public Run payload plus Temporal execution IDs.
type RunRecord struct {
	Run                api.Run
	TemporalWorkflowID string
	TemporalRunID      string
}

// Store persists graphs and runs across process restarts.
type Store interface {
	PutGraph(ctx context.Context, graph api.Graph) error
	GetGraph(ctx context.Context, id openapi_types.UUID) (api.Graph, bool, error)
	PutRun(ctx context.Context, record RunRecord) error
	GetRun(ctx context.Context, id openapi_types.UUID) (RunRecord, bool, error)
	UpdateRun(ctx context.Context, record RunRecord) error
	Close() error
}
