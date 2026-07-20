package server

import (
	"context"
	"errors"
	"strings"

	enums "go.temporal.io/api/enums/v1"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/internal/store"
	"github.com/madmmas/temflowral/backend/internal/temporal"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

// GraphRunner starts and inspects graph workflows.
type GraphRunner interface {
	StartGraphWorkflow(
		ctx context.Context,
		workflowID string,
		input temporal.GraphWorkflowInput,
	) (temporal.WorkflowExecution, error)
	DescribeGraphWorkflow(
		ctx context.Context,
		execution temporal.WorkflowExecution,
	) (temporal.WorkflowStatus, error)
}

// API implements the generated strict server interface.
type API struct {
	store    store.Store
	runner   GraphRunner
	registry *nodetype.Registry
}

var _ api.StrictServerInterface = (*API)(nil)

// NewAPI returns the HTTP API implementation. When registry is nil, the
// process-wide Temporal registry (built-ins) is used for ListNodeTypes.
func NewAPI(graphStore store.Store, runner GraphRunner, registry *nodetype.Registry) *API {
	if registry == nil {
		registry = temporal.CurrentRegistry()
	}
	return &API{store: graphStore, runner: runner, registry: registry}
}

func (apiServer *API) CreateGraph(
	ctx context.Context,
	request api.CreateGraphRequestObject,
) (api.CreateGraphResponseObject, error) {
	if request.Body == nil {
		return api.CreateGraph400JSONResponse{BadRequestJSONResponse: badRequest("request body is required")}, nil
	}

	now := nowUTC()
	graph := api.Graph{
		Id:        newGraphID(),
		Name:      request.Body.Name,
		Nodes:     request.Body.Nodes,
		Edges:     request.Body.Edges,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if graph.Nodes == nil {
		graph.Nodes = []api.Node{}
	}
	if graph.Edges == nil {
		graph.Edges = []api.Edge{}
	}
	for _, node := range graph.Nodes {
		if err := temporal.ValidateNodeConfig(node); err != nil {
			return api.CreateGraph400JSONResponse{
				BadRequestJSONResponse: badRequest(err.Error()),
			}, nil
		}
	}

	if err := apiServer.store.PutGraph(ctx, graph); err != nil {
		return api.CreateGraph500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	return api.CreateGraph201JSONResponse(graph), nil
}

func (apiServer *API) GetGraph(
	ctx context.Context,
	request api.GetGraphRequestObject,
) (api.GetGraphResponseObject, error) {
	graph, ok, err := apiServer.store.GetGraph(ctx, request.GraphId)
	if err != nil {
		return api.GetGraph500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	if !ok {
		return api.GetGraph404JSONResponse{NotFoundJSONResponse: notFound("graph not found")}, nil
	}
	return api.GetGraph200JSONResponse(graph), nil
}

func (apiServer *API) StartGraphRun(
	ctx context.Context,
	request api.StartGraphRunRequestObject,
) (api.StartGraphRunResponseObject, error) {
	graph, ok, err := apiServer.store.GetGraph(ctx, request.GraphId)
	if err != nil {
		return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	if !ok {
		return api.StartGraphRun404JSONResponse{NotFoundJSONResponse: notFound("graph not found")}, nil
	}

	if _, err := temporal.BuildExecutionPlan(graph); err != nil {
		return api.StartGraphRun409JSONResponse(conflict(err.Error())), nil
	}

	var workflowInput map[string]interface{}
	var idempotencyKey *string
	if request.Body != nil {
		if request.Body.Input != nil {
			workflowInput = *request.Body.Input
		}
		if request.Body.IdempotencyKey != nil {
			key := strings.TrimSpace(*request.Body.IdempotencyKey)
			if key == "" {
				return api.StartGraphRun400JSONResponse{
					BadRequestJSONResponse: badRequest("idempotencyKey must not be blank"),
				}, nil
			}
			if len(key) > 128 {
				return api.StartGraphRun400JSONResponse{
					BadRequestJSONResponse: badRequest("idempotencyKey must be at most 128 characters"),
				}, nil
			}
			idempotencyKey = &key
			existing, found, lookupErr := apiServer.store.GetRunByIdempotencyKey(ctx, graph.Id, key)
			if lookupErr != nil {
				return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: internalError(lookupErr.Error())}, nil
			}
			if found {
				return api.StartGraphRun202JSONResponse(existing.Run), nil
			}
		}
	}

	runID := newRunID()
	startedAt := nowUTC()
	execution, err := apiServer.runner.StartGraphWorkflow(ctx, runID.String(), temporal.GraphWorkflowInput{
		Graph: graph,
		Input: workflowInput,
	})
	if err != nil {
		return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}

	run := api.Run{
		Id:        runID,
		GraphId:   graph.Id,
		Status:    api.Running,
		StartedAt: startedAt,
	}
	if err := apiServer.store.PutRun(ctx, store.RunRecord{
		Run:                run,
		TemporalWorkflowID: execution.ID,
		TemporalRunID:      execution.RunID,
		IdempotencyKey:     idempotencyKey,
	}); err != nil {
		if errors.Is(err, store.ErrDuplicateIdempotencyKey) && idempotencyKey != nil {
			existing, found, lookupErr := apiServer.store.GetRunByIdempotencyKey(ctx, graph.Id, *idempotencyKey)
			if lookupErr != nil {
				return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: internalError(lookupErr.Error())}, nil
			}
			if found {
				return api.StartGraphRun202JSONResponse(existing.Run), nil
			}
		}
		return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	return api.StartGraphRun202JSONResponse(run), nil
}

func (apiServer *API) ListNodeTypes(
	_ context.Context,
	_ api.ListNodeTypesRequestObject,
) (api.ListNodeTypesResponseObject, error) {
	defs := apiServer.registry.List()
	nodeTypes := make([]api.NodeType, 0, len(defs))
	for _, def := range defs {
		nodeType := api.NodeType{
			Id:           def.ID,
			Name:         def.Name,
			ConfigSchema: def.ConfigSchema,
		}
		if def.Description != "" {
			description := def.Description
			nodeType.Description = &description
		}
		if def.Category != "" {
			category := def.Category
			nodeType.Category = &category
		}
		if len(def.OutputHandles) > 0 {
			handles := make([]api.NodeOutputHandle, 0, len(def.OutputHandles))
			for _, handle := range def.OutputHandles {
				item := api.NodeOutputHandle{Id: handle.ID}
				if handle.Label != "" {
					label := handle.Label
					item.Label = &label
				}
				handles = append(handles, item)
			}
			nodeType.OutputHandles = &handles
		}
		if def.OutputHandlesFromConfig != nil {
			nodeType.OutputHandlesFromConfig = &api.OutputHandlesFromConfig{
				Path: def.OutputHandlesFromConfig.Path,
			}
		}
		nodeTypes = append(nodeTypes, nodeType)
	}
	return api.ListNodeTypes200JSONResponse{NodeTypes: nodeTypes}, nil
}

func (apiServer *API) GetRun(
	ctx context.Context,
	request api.GetRunRequestObject,
) (api.GetRunResponseObject, error) {
	record, ok, err := apiServer.store.GetRun(ctx, request.RunId)
	if err != nil {
		return api.GetRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	if !ok {
		return api.GetRun404JSONResponse{NotFoundJSONResponse: notFound("run not found")}, nil
	}

	status, err := apiServer.runner.DescribeGraphWorkflow(ctx, temporal.WorkflowExecution{
		ID:    record.TemporalWorkflowID,
		RunID: record.TemporalRunID,
	})
	if err != nil {
		return api.GetRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}

	record.Run.Status = mapTemporalStatus(status.Status)
	switch record.Run.Status {
	case api.Completed:
		completedAt := nowUTC()
		record.Run.CompletedAt = &completedAt
		record.Run.Error = nil
		if status.Result != nil {
			result := graphResultToMap(*status.Result)
			record.Run.Result = &result
		}
	case api.Failed, api.Cancelled:
		completedAt := nowUTC()
		record.Run.CompletedAt = &completedAt
		if status.Error != "" {
			errMessage := status.Error
			record.Run.Error = &errMessage
		}
		record.Run.Result = nil
	default:
		record.Run.CompletedAt = nil
		record.Run.Error = nil
		record.Run.Result = nil
	}

	if err := apiServer.store.UpdateRun(ctx, record); err != nil {
		return api.GetRun500JSONResponse{InternalErrorJSONResponse: internalError(err.Error())}, nil
	}
	return api.GetRun200JSONResponse(record.Run), nil
}

func mapTemporalStatus(status enums.WorkflowExecutionStatus) api.RunStatus {
	switch status {
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return api.Completed
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
		enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return api.Failed
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return api.Cancelled
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
		enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return api.Running
	default:
		return api.Pending
	}
}

func graphResultToMap(result temporal.GraphWorkflowResult) map[string]interface{} {
	nodes := make([]map[string]interface{}, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		nodes = append(nodes, map[string]interface{}{
			"nodeId": node.NodeID,
			"value":  node.Value,
		})
	}
	return map[string]interface{}{"nodes": nodes}
}

func badRequest(message string) api.BadRequestJSONResponse {
	code := "bad_request"
	return api.BadRequestJSONResponse{Code: &code, Message: message}
}

func notFound(message string) api.NotFoundJSONResponse {
	code := "not_found"
	return api.NotFoundJSONResponse{Code: &code, Message: message}
}

func conflict(message string) api.Error {
	code := "conflict"
	return api.Error{Code: &code, Message: message}
}

func internalError(message string) api.InternalErrorJSONResponse {
	code := "internal_error"
	return api.InternalErrorJSONResponse{Code: &code, Message: message}
}
