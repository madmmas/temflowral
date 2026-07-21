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
	queryFn    func(context.Context, temporal.WorkflowExecution) (temporal.CurrentWait, error)
	signalFn   func(context.Context, temporal.WorkflowExecution, string, interface{}) error
}

func (stub *stubRunner) StartGraphWorkflow(
	ctx context.Context,
	workflowID string,
	input temporal.GraphWorkflowInput,
) (temporal.WorkflowExecution, error) {
	if stub.startFn == nil {
		return temporal.WorkflowExecution{}, fmt.Errorf("startFn not set")
	}
	return stub.startFn(ctx, workflowID, input)
}

func (stub *stubRunner) DescribeGraphWorkflow(
	ctx context.Context,
	execution temporal.WorkflowExecution,
) (temporal.WorkflowStatus, error) {
	if stub.describeFn == nil {
		return temporal.WorkflowStatus{}, fmt.Errorf("describeFn not set")
	}
	return stub.describeFn(ctx, execution)
}

func (stub *stubRunner) QueryCurrentWait(
	ctx context.Context,
	execution temporal.WorkflowExecution,
) (temporal.CurrentWait, error) {
	if stub.queryFn == nil {
		return temporal.CurrentWait{}, fmt.Errorf("queryFn not set")
	}
	return stub.queryFn(ctx, execution)
}

func (stub *stubRunner) SignalGraphWorkflow(
	ctx context.Context,
	execution temporal.WorkflowExecution,
	signalName string,
	payload interface{},
) error {
	if stub.signalFn == nil {
		return fmt.Errorf("signalFn not set")
	}
	return stub.signalFn(ctx, execution, signalName, payload)
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

func TestCreateGraphRejectsTaskQueueOnDelay(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"delay-1","type":"delay","position":{"x":100,"y":0},"config":{"seconds":5},"taskQueue":"special.queue"}
			],
			"edges":[{"id":"e1","source":"start-1","target":"delay-1"}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "taskQueue") {
		t.Fatalf("body = %s, want taskQueue validation error", recorder.Body.String())
	}
}

func TestCreateGraphAcceptsTaskQueueOnNoop(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"noop-1","type":"noop","position":{"x":100,"y":0},"taskQueue":"worker.gpu"}
			],
			"edges":[{"id":"e1","source":"start-1","target":"noop-1"}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	var created api.Graph
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if len(created.Nodes) != 2 || created.Nodes[1].TaskQueue == nil || *created.Nodes[1].TaskQueue != "worker.gpu" {
		t.Fatalf("created nodes = %#v, want taskQueue on noop", created.Nodes)
	}
}

func TestCreateGraphRejectsActivityOptionsOnDelay(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"delay-1","type":"delay","position":{"x":100,"y":0},"config":{"seconds":5},"activityOptions":{"startToCloseTimeoutSeconds":60}}
			],
			"edges":[{"id":"e1","source":"start-1","target":"delay-1"}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "activityOptions") {
		t.Fatalf("body = %s, want activityOptions validation error", recorder.Body.String())
	}
}

func TestCreateGraphAcceptsActivityOptionsOnNoop(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"noop-1","type":"noop","position":{"x":100,"y":0},"activityOptions":{"startToCloseTimeoutSeconds":60,"retryPolicy":{"maximumAttempts":2}}}
			],
			"edges":[{"id":"e1","source":"start-1","target":"noop-1"}]
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	var created api.Graph
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if len(created.Nodes) != 2 || created.Nodes[1].ActivityOptions == nil {
		t.Fatalf("created nodes = %#v, want activityOptions on noop", created.Nodes)
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
		!strings.Contains(body, `"id":"wait"`) ||
		!strings.Contains(body, `"id":"childWorkflow"`) {
		t.Fatalf("body = %s, want start, noop, http, delay, condition, wait, and childWorkflow node types", body)
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

	childType, ok := byID[temporal.ChildWorkflowNodeType]
	if !ok {
		t.Fatal("childWorkflow node type not found")
	}
	childProps, ok := childType.ConfigSchema["properties"].(map[string]interface{})
	if !ok || childProps["graph"] == nil {
		t.Errorf("childWorkflow config properties = %#v, want graph", childType.ConfigSchema["properties"])
	}
}

func TestSignalRunForwardsMatchingSignal(t *testing.T) {
	t.Parallel()

	var signaled struct {
		execution temporal.WorkflowExecution
		name      string
		payload   interface{}
		count     int
	}
	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			return temporal.WorkflowExecution{ID: workflowID, RunID: "temporal-run-1"}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING}, nil
		},
		queryFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.CurrentWait, error) {
			return temporal.CurrentWait{NodeID: "wait-1", Signal: "approval.granted"}, nil
		},
		signalFn: func(_ context.Context, execution temporal.WorkflowExecution, name string, payload interface{}) error {
			signaled.execution = execution
			signaled.name = name
			signaled.payload = payload
			signaled.count++
			return nil
		},
	}
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), runner, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"wait-1","type":"wait","position":{"x":100,"y":0},"config":{"signal":"approval.granted","timeoutSeconds":60}},
				{"id":"noop-received","type":"noop","position":{"x":200,"y":0}},
				{"id":"noop-timeout","type":"noop","position":{"x":200,"y":100}}
			],
			"edges":[
				{"id":"e0","source":"start-1","target":"wait-1"},
				{"id":"e-recv","source":"wait-1","target":"noop-received","sourceHandle":"received"},
				{"id":"e-to","source":"wait-1","target":"noop-timeout","sourceHandle":"timedOut"}
			]
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

	runRequest := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(`{}`))
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	if runRecorder.Code != http.StatusAccepted {
		t.Fatalf("run status = %d body=%s", runRecorder.Code, runRecorder.Body.String())
	}
	var started api.Run
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	signalRequest := httptest.NewRequest(
		http.MethodPost,
		"/runs/"+started.Id.String()+"/signal",
		strings.NewReader(`{"signal":"approval.granted","payload":{"approvedBy":"alice"}}`),
	)
	signalRequest.Header.Set("Content-Type", "application/json")
	signalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(signalRecorder, signalRequest)
	if signalRecorder.Code != http.StatusAccepted {
		t.Fatalf("signal status = %d body=%s", signalRecorder.Code, signalRecorder.Body.String())
	}

	var accepted api.SignalRunResponse
	if err := json.Unmarshal(signalRecorder.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode signal response: %v", err)
	}
	if accepted.RunId != started.Id || accepted.Signal != "approval.granted" {
		t.Fatalf("accepted = %#v, want run %s signal approval.granted", accepted, started.Id)
	}
	if signaled.count != 1 || signaled.name != "approval.granted" || signaled.execution.ID != started.Id.String() {
		t.Fatalf("signaled = %#v, want one signal to workflow %s", signaled, started.Id)
	}
	payload, ok := signaled.payload.(map[string]interface{})
	if !ok || payload["approvedBy"] != "alice" {
		t.Fatalf("payload = %#v, want approvedBy=alice", signaled.payload)
	}
}

func TestSignalRunRejectsUnknownRun(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/runs/550e8400-e29b-41d4-a716-446655440000/signal",
		strings.NewReader(`{"signal":"approval.granted"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestSignalRunRejectsInvalidSignalName(t *testing.T) {
	t.Parallel()

	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), &stubRunner{}, nil))
	request := httptest.NewRequest(
		http.MethodPost,
		"/runs/550e8400-e29b-41d4-a716-446655440000/signal",
		strings.NewReader(`{"signal":"not a valid signal"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}

func TestSignalRunConflictWhenNotWaiting(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			return temporal.WorkflowExecution{ID: workflowID, RunID: "temporal-run-1"}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING}, nil
		},
		queryFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.CurrentWait, error) {
			return temporal.CurrentWait{}, nil
		},
		signalFn: func(context.Context, temporal.WorkflowExecution, string, interface{}) error {
			t.Fatal("SignalGraphWorkflow should not be called")
			return nil
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
	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	runRequest := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(`{}`))
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	var started api.Run
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	signalRequest := httptest.NewRequest(
		http.MethodPost,
		"/runs/"+started.Id.String()+"/signal",
		strings.NewReader(`{"signal":"approval.granted"}`),
	)
	signalRequest.Header.Set("Content-Type", "application/json")
	signalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(signalRecorder, signalRequest)
	if signalRecorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d body=%s", signalRecorder.Code, http.StatusConflict, signalRecorder.Body.String())
	}
	if !strings.Contains(signalRecorder.Body.String(), "not currently waiting") {
		t.Fatalf("body = %s, want not currently waiting", signalRecorder.Body.String())
	}
}

func TestSignalRunConflictWhenWrongSignal(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			return temporal.WorkflowExecution{ID: workflowID, RunID: "temporal-run-1"}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{Status: enums.WORKFLOW_EXECUTION_STATUS_RUNNING}, nil
		},
		queryFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.CurrentWait, error) {
			return temporal.CurrentWait{NodeID: "wait-1", Signal: "approval.granted"}, nil
		},
		signalFn: func(context.Context, temporal.WorkflowExecution, string, interface{}) error {
			t.Fatal("SignalGraphWorkflow should not be called")
			return nil
		},
	}
	handler := NewHandler([]byte("openapi: 3.1.0\n"), NewAPI(store.NewMemoryStore(), runner, nil))

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/graphs",
		strings.NewReader(`{
			"nodes":[
				{"id":"start-1","type":"start","position":{"x":0,"y":0}},
				{"id":"wait-1","type":"wait","position":{"x":100,"y":0},"config":{"signal":"approval.granted","timeoutSeconds":60}},
				{"id":"noop-received","type":"noop","position":{"x":200,"y":0}},
				{"id":"noop-timeout","type":"noop","position":{"x":200,"y":100}}
			],
			"edges":[
				{"id":"e0","source":"start-1","target":"wait-1"},
				{"id":"e-recv","source":"wait-1","target":"noop-received","sourceHandle":"received"},
				{"id":"e-to","source":"wait-1","target":"noop-timeout","sourceHandle":"timedOut"}
			]
		}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createRecorder, createRequest)
	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	runRequest := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(`{}`))
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	var started api.Run
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	signalRequest := httptest.NewRequest(
		http.MethodPost,
		"/runs/"+started.Id.String()+"/signal",
		strings.NewReader(`{"signal":"other.signal"}`),
	)
	signalRequest.Header.Set("Content-Type", "application/json")
	signalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(signalRecorder, signalRequest)
	if signalRecorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d body=%s", signalRecorder.Code, http.StatusConflict, signalRecorder.Body.String())
	}
}

func TestSignalRunConflictWhenNotRunning(t *testing.T) {
	t.Parallel()

	runner := &stubRunner{
		startFn: func(_ context.Context, workflowID string, _ temporal.GraphWorkflowInput) (temporal.WorkflowExecution, error) {
			return temporal.WorkflowExecution{ID: workflowID, RunID: "temporal-run-1"}, nil
		},
		describeFn: func(_ context.Context, _ temporal.WorkflowExecution) (temporal.WorkflowStatus, error) {
			return temporal.WorkflowStatus{Status: enums.WORKFLOW_EXECUTION_STATUS_COMPLETED}, nil
		},
		queryFn: func(context.Context, temporal.WorkflowExecution) (temporal.CurrentWait, error) {
			t.Fatal("QueryCurrentWait should not be called")
			return temporal.CurrentWait{}, nil
		},
		signalFn: func(context.Context, temporal.WorkflowExecution, string, interface{}) error {
			t.Fatal("SignalGraphWorkflow should not be called")
			return nil
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
	var created api.Graph
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	runRequest := httptest.NewRequest(http.MethodPost, "/graphs/"+created.Id.String()+"/run", strings.NewReader(`{}`))
	runRequest.Header.Set("Content-Type", "application/json")
	runRecorder := httptest.NewRecorder()
	handler.ServeHTTP(runRecorder, runRequest)
	var started api.Run
	if err := json.Unmarshal(runRecorder.Body.Bytes(), &started); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	signalRequest := httptest.NewRequest(
		http.MethodPost,
		"/runs/"+started.Id.String()+"/signal",
		strings.NewReader(`{"signal":"approval.granted"}`),
	)
	signalRequest.Header.Set("Content-Type", "application/json")
	signalRecorder := httptest.NewRecorder()
	handler.ServeHTTP(signalRecorder, signalRequest)
	if signalRecorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d body=%s", signalRecorder.Code, http.StatusConflict, signalRecorder.Body.String())
	}
}
