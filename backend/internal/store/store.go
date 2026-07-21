package store

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// ErrDuplicateIdempotencyKey is returned when PutRun conflicts on a graph's
// idempotency key. Callers should load the existing run and return it.
var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key for graph")

// RunRecord keeps the public Run payload plus Temporal execution IDs and an
// optional caller-supplied idempotency key.
type RunRecord struct {
	Run                api.Run
	TemporalWorkflowID string
	TemporalRunID      string
	IdempotencyKey     *string
}

// Store persists graphs and runs across process restarts.
type Store interface {
	PutGraph(ctx context.Context, graph api.Graph) error
	GetGraph(ctx context.Context, id openapi_types.UUID) (api.Graph, bool, error)
	PutRun(ctx context.Context, record RunRecord) error
	GetRun(ctx context.Context, id openapi_types.UUID) (RunRecord, bool, error)
	GetRunByIdempotencyKey(
		ctx context.Context,
		graphID openapi_types.UUID,
		key string,
	) (RunRecord, bool, error)
	UpdateRun(ctx context.Context, record RunRecord) error
	Close() error
}
