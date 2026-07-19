package server

import (
	"context"

	enums "go.temporal.io/api/enums/v1"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/internal/temporal"
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
	store  *Store
	runner GraphRunner
}

var _ api.StrictServerInterface = (*API)(nil)

// NewAPI returns the HTTP API implementation.
func NewAPI(store *Store, runner GraphRunner) *API {
	return &API{store: store, runner: runner}
}

func (apiServer *API) CreateGraph(
	_ context.Context,
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

	apiServer.store.PutGraph(graph)
	return api.CreateGraph201JSONResponse(graph), nil
}

func (apiServer *API) GetGraph(
	_ context.Context,
	request api.GetGraphRequestObject,
) (api.GetGraphResponseObject, error) {
	graph, ok := apiServer.store.GetGraph(request.GraphId)
	if !ok {
		return api.GetGraph404JSONResponse{NotFoundJSONResponse: notFound("graph not found")}, nil
	}
	return api.GetGraph200JSONResponse(graph), nil
}

func (apiServer *API) StartGraphRun(
	ctx context.Context,
	request api.StartGraphRunRequestObject,
) (api.StartGraphRunResponseObject, error) {
	graph, ok := apiServer.store.GetGraph(request.GraphId)
	if !ok {
		return api.StartGraphRun404JSONResponse{NotFoundJSONResponse: notFound("graph not found")}, nil
	}

	if _, err := temporal.BuildExecutionPlan(graph); err != nil {
		return api.StartGraphRun409JSONResponse(conflict(err.Error())), nil
	}

	var workflowInput map[string]interface{}
	if request.Body != nil && request.Body.Input != nil {
		workflowInput = *request.Body.Input
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
	apiServer.store.PutRun(RunRecord{
		Run:                run,
		TemporalWorkflowID: execution.ID,
		TemporalRunID:      execution.RunID,
	})
	return api.StartGraphRun202JSONResponse(run), nil
}

func (apiServer *API) ListNodeTypes(
	_ context.Context,
	_ api.ListNodeTypesRequestObject,
) (api.ListNodeTypesResponseObject, error) {
	core := "core"
	startDescription := "Workflow entry point"
	noopDescription := "No-op activity used to smoke-test graph execution"
	return api.ListNodeTypes200JSONResponse{
		NodeTypes: []api.NodeType{
			{
				Id:          temporal.StartNodeType,
				Name:        "Start",
				Description: &startDescription,
				Category:    &core,
				ConfigSchema: map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
				},
			},
			{
				Id:          temporal.NoopNodeType,
				Name:        "No-op",
				Description: &noopDescription,
				Category:    &core,
				ConfigSchema: map[string]interface{}{
					"type":                 "object",
					"additionalProperties": true,
				},
			},
		},
	}, nil
}

func (apiServer *API) GetRun(
	ctx context.Context,
	request api.GetRunRequestObject,
) (api.GetRunResponseObject, error) {
	record, ok := apiServer.store.GetRun(request.RunId)
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

	apiServer.store.UpdateRun(record)
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
