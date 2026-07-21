package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	enums "go.temporal.io/api/enums/v1"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/internal/store"
	"github.com/madmmas/temflowral/backend/internal/temporal"
)

type stubRunner struct {
	startFn    func(context.Context, string, temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error)
	describeFn func(context.Context, temporal.WorkflowExecution) (temporal.WorkflowStatus, error)
}

func (stub *stubRunner) StartGraphWorkflow(
	ctx context.Context,
	workflowID string,
	input temporal.GraphWorkflowInput,
) (temporal.WorkflowExecution, error) {
	return stub.startFn(ctx, workflowID, input)
}

func (stub *stubRunner) DescribeGraphWorkflow(
	ctx context.Context,
	execution temporal.WorkflowExecution,
) (temporal.WorkflowStatus, error) {
	return stub.describeFn(ctx, execution)
}

func TestCreateAndGetGraph(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"name":"demo",
			"nodes":[{"id":"start-1","type":"start","position":{"x":0,"y":0}}],
			"edges":[]
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", createRecorder.Code, http.StatusCreated, createRecorder.Body.String())
	}

	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Id.String() == "" {
		t.Fatal("created graph id is empty")
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/graphs/"+created.Id.String(), nil)
	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRecorder.Code, http.StatusOK)
	}
}

func TestCreateGraphRejectsInvalidHTTPConfig(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"http-1","type":"http","position":{"x":100,"y":0},"config":{"method":"GET","url":"file:///etc/passwd"}}
			],
			"edges":[{"id":"e1","source":"start-1","target":"http-1"}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "scheme must be http or https") {
		t.Fatalf("body = %s, want HTTP config validation error", recorder.Body.String())
	}
}

func TestStartGraphRunRejectsInvalidGraph(t *testing.T) {
	t.Parallel()

	graphStore := store.NewMemoryStore()
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(graphStore, &stubRunner{}, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[{"id":"noop-1","type":"noop","position":{"x":0,"y":0}}],
			"edges":[]
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRecorder.Code, http.StatusCreated)
	}

	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	runRequest := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(`{}`))
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	if runRecorder.Code != http.StatusConflict {
		t.Fatalf("run status = %d, want %d body=%s", runRecorder.Code, http.StatusConflict, runRecorder.Body.String())
	}
}

func TestStartAndGetGraphRun(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			return temporal.WorkflowExecution{ID: workflowID, RunID: "run-1"}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{
				Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				Result: &temporal.GraphWorkflowResult{
					Nodes: []temporal.NodeResult{
						{NodeID: "start-1", Value: map[string]interface{}{"message": "hello"}},
						{NodeID: "noop-1", Value: map[string]interface{}{"type": "noop"}},
					},
				},
			}, nil
		},
	}
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), runner, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"noop-1","type":"noop","position":{"x":100,"y":0}}
			],
			"edges":[{"id":"e1","source":"start-1","target":"noop-1"}]
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	runRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs/"+created.Id.String()+"/run",
		strings.NewReader(`{"input":{"message":"hello"}}`),
	)
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	if runRecorder.Code != http.StatusAccepted {
		t.Fatalf("run status = %d body=%s", runRecorder.Code, runRecorder.Body.String())
	}

	var started api.Run
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode run response: %v", err)
	}
	if started.Status != api.Running {
		t.Fatalf("started status = %q, want %q", started.Status, api.Running)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/runs/"+started.Id.String(), nil)
	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, getRequest)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get run status = %d body=%s", getRecorder.Code, getRecorder.Body.String())
	}

	body, err := io.ReadAll(getRecorder.Body)
	if err != nil {
		t.Fatalf("read get run body: %v", err)
	}
	if !strings.Contains(string(body), `"status":"completed"`) {
		t.Fatalf("get run body = %s, want completed status", body)
	}
}

func TestStartGraphRunIdempotencyKeyDedupes(t *testing.T) {
	t.Parallel()

	startCount := 0
	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			startCount++
			return temporal.WorkflowExecution{ID: workflowID, RunID: fmt.Sprintf("run-%d", startCount)}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING}, nil
		},
	}
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), runner, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"noop-1","type":"noop","position":{"x":100,"y":0}}
			],
			"edges":[{"id":"e1","source":"start-1","target":"noop-1"}]
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}
	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	body := `{"idempotencyKey":"delivery-1","input":{"message":"hello"}}`
	first := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(body))
	first.Header.Set("Content-Type", "application/json")
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, first)
	if firstRecorder.Code != http.StatusAccepted {
		t.Fatalf("first run status = %d body=%s", firstRecorder.Code, firstRecorder.Body.String())
	}
	var firstRun api.Run
	if err := json.Unmarshal(firstRecorder.Body.Bytes(), &firstRun); err != nil {
		t.Fatalf("decode first run: %v", err)
	}

	second := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(body))
	second.Header.Set("Content-Type", "application/json")
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, second)
	if secondRecorder.Code != http.StatusAccepted {
		t.Fatalf("second run status = %d body=%s", secondRecorder.Code, secondRecorder.Body.String())
	}
	var secondRun api.Run
	if err := json.Unmarshal(secondRecorder.Body.Bytes(), &secondRun); err != nil {
		t.Fatalf("decode second run: %v", err)
	}

	if firstRun.Id != secondRun.Id {
		t.Fatalf("run ids = %s vs %s, want same idempotent run", firstRun.Id, secondRun.Id)
	}
	if startCount != 1 {
		t.Fatalf("StartGraphWorkflow calls = %d, want 1", startCount)
	}
}

func TestListNodeTypes(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(http.MethodGet, "/node-types", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"id":"start"`) ||
		!strings.Contains(body, `"id":"noop"`) ||
		!strings.Contains(body, `"id":"http"`) ||
		!strings.Contains(body, `"id":"delay"`) ||
		!strings.Contains(body, `"id":"condition"`) ||
		!strings.Contains(body, `"id":"wait"`) {
		t.Fatalf("body = %s, want start, noop, http, delay, condition, and wait node types", body)
	}

	var registry api.NodeTypeList
	if err := json.Unmarshal(recorder.Body.Bytes(), &registry); err != nil {
		t.Fatalf("decode node type registry: %v", err)
	}
	byID := make(map[string]api.NodeType, len(registry.NodeTypes))
	for _, nodeType := range registry.NodeTypes {
		byID[nodeType.Id] = nodeType
	}

	httpType, ok := byID[temporal.HTTPNodeType]
	if !ok {
		t.Fatal("HTTP node type not found")
	}
	if httpType.ConfigSchema["additionalProperties"] != false {
		t.Errorf("HTTP config additionalProperties = %#v, want false", httpType.ConfigSchema["additionalProperties"])
	}
	httpProps, ok := httpType.ConfigSchema["properties"].(map[string]interface{})
	if !ok || httpProps["method"] == nil || httpProps["url"] == nil {
		t.Errorf("HTTP config properties = %#v, want method and url", httpType.ConfigSchema["properties"])
	}

	delayType, ok := byID[temporal.DelayNodeType]
	if !ok {
		t.Fatal("delay node type not found")
	}
	if delayType.ConfigSchema["additionalProperties"] != false {
		t.Errorf("delay config additionalProperties = %#v, want false", delayType.ConfigSchema["additionalProperties"])
	}
	delayProps, ok := delayType.ConfigSchema["properties"].(map[string]interface{})
	if !ok || delayProps["seconds"] == nil {
		t.Errorf("delay config properties = %#v, want seconds", delayType.ConfigSchema["properties"])
	}

	conditionType, ok := byID[temporal.ConditionNodeType]
	if !ok {
		t.Fatal("condition node type not found")
	}
	conditionProps, ok := conditionType.ConfigSchema["properties"].(map[string]interface{})
	if !ok || conditionProps["field"] == nil || conditionProps["equals"] == nil {
		t.Errorf("condition config properties = %#v, want field and equals", conditionType.ConfigSchema["properties"])
	}
	if conditionType.OutputHandles == nil || len(*conditionType.OutputHandles) != 2 {
		t.Fatalf("condition outputHandles = %#v, want true/false", conditionType.OutputHandles)
	}

	waitType, ok := byID[temporal.WaitNodeType]
	if !ok {
		t.Fatal("wait node type not found")
	}
	waitProps, ok := waitType.ConfigSchema["properties"].(map[string]interface{})
	if !ok || waitProps["signal"] == nil || waitProps["timeoutSeconds"] == nil {
		t.Errorf("wait config properties = %#v, want signal and timeoutSeconds", waitType.ConfigSchema["properties"])
	}
	if waitType.OutputHandles == nil || len(*waitType.OutputHandles) != 2 {
		t.Fatalf("wait outputHandles = %#v, want received/timedOut", waitType.OutputHandles)
	}
}
