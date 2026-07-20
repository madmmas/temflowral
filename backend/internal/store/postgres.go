package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// PostgresStore persists graphs and runs in PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// OpenPostgres connects to DATABASE_URL, ensures the database and schema exist,
// and returns a store.
func OpenPostgres(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	if err := ensureDatabase(ctx, databaseURL); err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	store := &PostgresStore{pool: pool}
	if err := store.ensureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func ensureDatabase(ctx context.Context, databaseURL string) error {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database URL: %w", err)
	}
	dbName := config.ConnConfig.Database
	if dbName == "" || dbName == "postgres" || dbName == "temporal" {
		return nil
	}
	if !isSafeDatabaseName(dbName) {
		return fmt.Errorf("database name %q contains unsupported characters", dbName)
	}

	adminConfig := config.Copy()
	adminConfig.ConnConfig.Database = "postgres"
	adminPool, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		return fmt.Errorf("connect to postgres admin database: %w", err)
	}
	defer adminPool.Close()

	var exists bool
	if err := adminPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check database %q: %w", dbName, err)
	}
	if exists {
		return nil
	}
	// Database names cannot be parameterized; Identifier quotes safely.
	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
		return fmt.Errorf("create database %q: %w", dbName, err)
	}
	return nil
}

func isSafeDatabaseName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}

func (store *PostgresStore) ensureSchema(ctx context.Context) error {
	if _, err := store.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}
	return nil
}

func (store *PostgresStore) Close() error {
	store.pool.Close()
	return nil
}

func (store *PostgresStore) PutGraph(ctx context.Context, graph api.Graph) error {
	nodes, err := json.Marshal(graph.Nodes)
	if err != nil {
		return fmt.Errorf("encode graph nodes: %w", err)
	}
	edges, err := json.Marshal(graph.Edges)
	if err != nil {
		return fmt.Errorf("encode graph edges: %w", err)
	}
	_, err = store.pool.Exec(ctx, `
		INSERT INTO graphs (id, name, nodes, edges, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			nodes = EXCLUDED.nodes,
			edges = EXCLUDED.edges,
			updated_at = EXCLUDED.updated_at
	`, graph.Id, graph.Name, nodes, edges, graph.CreatedAt.UTC(), graph.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("put graph %s: %w", graph.Id, err)
	}
	return nil
}

func (store *PostgresStore) GetGraph(ctx context.Context, id openapi_types.UUID) (api.Graph, bool, error) {
	var (
		graph api.Graph
		name  *string
		nodes []byte
		edges []byte
	)
	err := store.pool.QueryRow(ctx, `
		SELECT id, name, nodes, edges, created_at, updated_at
		FROM graphs WHERE id = $1
	`, id).Scan(&graph.Id, &name, &nodes, &edges, &graph.CreatedAt, &graph.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return api.Graph{}, false, nil
	}
	if err != nil {
		return api.Graph{}, false, fmt.Errorf("get graph %s: %w", id, err)
	}
	graph.Name = name
	if err := json.Unmarshal(nodes, &graph.Nodes); err != nil {
		return api.Graph{}, false, fmt.Errorf("decode graph nodes: %w", err)
	}
	if err := json.Unmarshal(edges, &graph.Edges); err != nil {
		return api.Graph{}, false, fmt.Errorf("decode graph edges: %w", err)
	}
	if graph.Nodes == nil {
		graph.Nodes = []api.Node{}
	}
	if graph.Edges == nil {
		graph.Edges = []api.Edge{}
	}
	return graph, true, nil
}

func (store *PostgresStore) PutRun(ctx context.Context, record RunRecord) error {
	return store.upsertRun(ctx, record, false)
}

func (store *PostgresStore) UpdateRun(ctx context.Context, record RunRecord) error {
	return store.upsertRun(ctx, record, true)
}

func (store *PostgresStore) upsertRun(ctx context.Context, record RunRecord, requireExisting bool) error {
	resultJSON, err := encodeOptionalJSON(record.Run.Result)
	if err != nil {
		return fmt.Errorf("encode run result: %w", err)
	}
	var completedAt *time.Time
	if record.Run.CompletedAt != nil {
		utc := record.Run.CompletedAt.UTC()
		completedAt = &utc
	}

	if requireExisting {
		tag, err := store.pool.Exec(ctx, `
			UPDATE runs SET
				graph_id = $2,
				status = $3,
				started_at = $4,
				completed_at = $5,
				result = $6,
				error = $7,
				temporal_workflow_id = $8,
				temporal_run_id = $9
			WHERE id = $1
		`,
			record.Run.Id,
			record.Run.GraphId,
			string(record.Run.Status),
			record.Run.StartedAt.UTC(),
			completedAt,
			resultJSON,
			record.Run.Error,
			record.TemporalWorkflowID,
			record.TemporalRunID,
		)
		if err != nil {
			return fmt.Errorf("update run %s: %w", record.Run.Id, err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("update run %s: not found", record.Run.Id)
		}
		return nil
	}

	_, err = store.pool.Exec(ctx, `
		INSERT INTO runs (
			id, graph_id, status, started_at, completed_at, result, error,
			temporal_workflow_id, temporal_run_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		record.Run.Id,
		record.Run.GraphId,
		string(record.Run.Status),
		record.Run.StartedAt.UTC(),
		completedAt,
		resultJSON,
		record.Run.Error,
		record.TemporalWorkflowID,
		record.TemporalRunID,
	)
	if err != nil {
		return fmt.Errorf("put run %s: %w", record.Run.Id, err)
	}
	return nil
}

func (store *PostgresStore) GetRun(ctx context.Context, id openapi_types.UUID) (RunRecord, bool, error) {
	var (
		record      RunRecord
		status      string
		resultJSON  []byte
		completedAt *time.Time
	)
	err := store.pool.QueryRow(ctx, `
		SELECT id, graph_id, status, started_at, completed_at, result, error,
		       temporal_workflow_id, temporal_run_id
		FROM runs WHERE id = $1
	`, id).Scan(
		&record.Run.Id,
		&record.Run.GraphId,
		&status,
		&record.Run.StartedAt,
		&completedAt,
		&resultJSON,
		&record.Run.Error,
		&record.TemporalWorkflowID,
		&record.TemporalRunID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return RunRecord{}, false, nil
	}
	if err != nil {
		return RunRecord{}, false, fmt.Errorf("get run %s: %w", id, err)
	}
	record.Run.Status = api.RunStatus(status)
	record.Run.CompletedAt = completedAt
	if len(resultJSON) > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return RunRecord{}, false, fmt.Errorf("decode run result: %w", err)
		}
		record.Run.Result = &result
	}
	return record, true, nil
}

func encodeOptionalJSON(value *map[string]interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}
