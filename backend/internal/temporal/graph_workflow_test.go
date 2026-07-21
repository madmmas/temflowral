package temporal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestGraphWorkflowExecutesNoopNodes(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	input := GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "noop-1", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e1", Source: "start-1", Target: "noop-1"},
			},
		},
		Input: map[string]interface{}{"message": "hello"},
	}

	environment.ExecuteWorkflow(GraphWorkflow, input)
	if !environment.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("result nodes = %d, want 2", len(result.Nodes))
	}
	if result.Nodes[0].NodeID != "start-1" {
		t.Errorf("first node = %q, want start-1", result.Nodes[0].NodeID)
	}
	if got := result.Nodes[0].Value["message"]; got != "hello" {
		t.Errorf("start value message = %#v, want hello", got)
	}
	if result.Nodes[1].NodeID != "noop-1" {
		t.Errorf("second node = %q, want noop-1", result.Nodes[1].NodeID)
	}
}

func TestGraphWorkflowRunsDelayNodeTimer(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()

	// The delay node uses a durable workflow timer; the test environment
	// auto-advances simulated time, so assert a timer actually fired.
	timerFired := false
	environment.SetOnTimerFiredListener(func(string) {
		timerFired = true
	})

	config := map[string]interface{}{"seconds": 30}
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "delay-1", Type: DelayNodeType, Config: &config},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "delay-1"}},
		},
	})
	if !timerFired {
		t.Fatal("expected a durable timer to fire for the delay node")
	}
	if !environment.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if len(result.Nodes) != 2 || result.Nodes[1].NodeID != "delay-1" {
		t.Fatalf("result nodes = %#v, want delay node result", result.Nodes)
	}
	if got := result.Nodes[1].Value["seconds"]; got != float64(30) {
		t.Errorf("delay seconds = %#v, want 30", got)
	}
}

func TestGraphWorkflowTakesTrueBranch(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	config := map[string]interface{}{"field": "status", "equals": "ok"}
	trueHandle := ConditionTrueHandle
	falseHandle := ConditionFalseHandle
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "cond-1", Type: ConditionNodeType, Config: &config},
				{Id: "noop-true", Type: NoopNodeType},
				{Id: "noop-false", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "cond-1"},
				{Id: "e-true", Source: "cond-1", Target: "noop-true", SourceHandle: &trueHandle},
				{Id: "e-false", Source: "cond-1", Target: "noop-false", SourceHandle: &falseHandle},
			},
		},
		Input: map[string]interface{}{"status": "ok"},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "cond-1", "noop-true"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
	if got := result.Nodes[1].Value["branch"]; got != ConditionTrueHandle {
		t.Errorf("branch = %#v, want %q", got, ConditionTrueHandle)
	}
}

func TestGraphWorkflowTakesFalseBranch(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	config := map[string]interface{}{"field": "status", "equals": "ok"}
	trueHandle := ConditionTrueHandle
	falseHandle := ConditionFalseHandle
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "cond-1", Type: ConditionNodeType, Config: &config},
				{Id: "noop-true", Type: NoopNodeType},
				{Id: "noop-false", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "cond-1"},
				{Id: "e-true", Source: "cond-1", Target: "noop-true", SourceHandle: &trueHandle},
				{Id: "e-false", Source: "cond-1", Target: "noop-false", SourceHandle: &falseHandle},
			},
		},
		Input: map[string]interface{}{"status": "fail"},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "cond-1", "noop-false"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
}

func TestGraphWorkflowJoinAfterTakenBranch(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	config := map[string]interface{}{"field": "status", "equals": "ok"}
	trueHandle := ConditionTrueHandle
	falseHandle := ConditionFalseHandle
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "cond-1", Type: ConditionNodeType, Config: &config},
				{Id: "noop-true", Type: NoopNodeType},
				{Id: "noop-false", Type: NoopNodeType},
				{Id: "join", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "cond-1"},
				{Id: "e-true", Source: "cond-1", Target: "noop-true", SourceHandle: &trueHandle},
				{Id: "e-false", Source: "cond-1", Target: "noop-false", SourceHandle: &falseHandle},
				{Id: "e-join-t", Source: "noop-true", Target: "join"},
				{Id: "e-join-f", Source: "noop-false", Target: "join"},
			},
		},
		Input: map[string]interface{}{"status": "ok"},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "cond-1", "noop-true", "join"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
}

func TestGraphWorkflowResolvesHTTPConfigTemplates(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()

	var gotURL string
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			if input.Node.Config != nil {
				if url, ok := (*input.Node.Config)["url"].(string); ok {
					gotURL = url
				}
			}
			return NodeResult{
				NodeID: input.Node.ID,
				Value:  map[string]interface{}{"statusCode": 200, "body": "ok"},
			}, nil
		},
		activity.RegisterOptions{Name: HTTPNodeActivityName},
	)

	config := map[string]interface{}{
		"method": "GET",
		"url":    "https://api.example.com/items/{{ nodes.start-1.output.id }}",
	}
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "http-1", Type: HTTPNodeType, Config: &config},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "http-1"}},
		},
		Input: map[string]interface{}{"id": "abc"},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}
	if gotURL != "https://api.example.com/items/abc" {
		t.Fatalf("resolved url = %q, want https://api.example.com/items/abc", gotURL)
	}
}

func TestGraphWorkflowRunsChildWorkflow(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterWorkflow(GraphWorkflow)
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	childGraph := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":       "start-1",
				"type":     "start",
				"position": map[string]interface{}{"x": 0.0, "y": 0.0},
			},
			map[string]interface{}{
				"id":       "noop-1",
				"type":     "noop",
				"position": map[string]interface{}{"x": 100.0, "y": 0.0},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "start-1", "target": "noop-1"},
		},
	}
	config := map[string]interface{}{
		"graph": childGraph,
		"input": map[string]interface{}{"message": "from-parent"},
	}

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "child-1", Type: ChildWorkflowNodeType, Config: &config},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "child-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if len(result.Nodes) != 2 || result.Nodes[1].NodeID != "child-1" {
		t.Fatalf("result = %#v, want child-1", result.Nodes)
	}
	if got := result.Nodes[1].Value["type"]; got != ChildWorkflowNodeType {
		t.Fatalf("type = %#v, want %q", got, ChildWorkflowNodeType)
	}
	nodes, ok := result.Nodes[1].Value["nodes"].([]interface{})
	if !ok || len(nodes) != 2 {
		t.Fatalf("child nodes = %#v, want 2 entries", result.Nodes[1].Value["nodes"])
	}
}

func TestGraphWorkflowChildWorkflowFailureFailsParent(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterWorkflow(GraphWorkflow)
	environment.RegisterActivityWithOptions(
		func(_ context.Context, _ NodeActivityInput) (NodeResult, error) {
			return NodeResult{}, fmt.Errorf("child activity failed")
		},
		activity.RegisterOptions{Name: NoopNodeActivityName},
	)

	childGraph := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":       "start-1",
				"type":     "start",
				"position": map[string]interface{}{"x": 0.0, "y": 0.0},
			},
			map[string]interface{}{
				"id":       "noop-1",
				"type":     "noop",
				"position": map[string]interface{}{"x": 100.0, "y": 0.0},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "start-1", "target": "noop-1"},
		},
	}
	config := map[string]interface{}{"graph": childGraph}

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "child-1", Type: ChildWorkflowNodeType, Config: &config},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "child-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err == nil {
		t.Fatal("workflow error = nil, want child failure to fail parent")
	}
}

func TestGraphWorkflowRoutesActivityToTaskQueue(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	const wantQueue = "worker.gpu"
	var gotQueue string
	environment.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, _ converter.EncodedValues) {
		gotQueue = info.TaskQueue
	})

	queue := wantQueue
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "noop-1", Type: NoopNodeType, TaskQueue: &queue},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "noop-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}
	if gotQueue != wantQueue {
		t.Fatalf("activity TaskQueue = %q, want %q", gotQueue, wantQueue)
	}
}

func TestGraphWorkflowAppliesPerNodeRetryPolicy(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()

	attempts := 0
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			attempts++
			if attempts < 3 {
				return NodeResult{}, fmt.Errorf("transient failure %d", attempts)
			}
			return NodeResult{
				NodeID: input.Node.ID,
				Value:  map[string]interface{}{"attempts": attempts},
			}, nil
		},
		activity.RegisterOptions{Name: NoopNodeActivityName},
	)

	options := &api.ActivityOptions{
		RetryPolicy: &api.RetryPolicy{MaximumAttempts: 3},
	}
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "noop-1", Type: NoopNodeType, ActivityOptions: options},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "noop-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("result = %#v, want 2 nodes", result.Nodes)
	}
	// Temporal payload decode uses JSON numbers → float64.
	if got := result.Nodes[1].Value["attempts"]; got != float64(3) {
		t.Fatalf("attempts = %#v, want 3", got)
	}
}

func TestGraphWorkflowDefaultRetryDoesNotRetry(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()

	attempts := 0
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			attempts++
			return NodeResult{}, fmt.Errorf("always fails")
		},
		activity.RegisterOptions{Name: NoopNodeActivityName},
	)

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "noop-1", Type: NoopNodeType},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "noop-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err == nil {
		t.Fatal("workflow error = nil, want an error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1 (default MaximumAttempts)", attempts)
	}
}

func TestGraphWorkflowDispatchesHTTPNode(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			return NodeResult{
				NodeID: input.Node.ID,
				Value:  map[string]interface{}{"statusCode": 200, "body": "ok"},
			}, nil
		},
		activity.RegisterOptions{Name: HTTPNodeActivityName},
	)

	config := map[string]interface{}{"method": "GET", "url": "https://api.example.com"}
	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "http-1", Type: HTTPNodeType, Config: &config},
			},
			Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "http-1"}},
		},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if len(result.Nodes) != 2 || result.Nodes[1].NodeID != "http-1" {
		t.Fatalf("result nodes = %#v, want HTTP node result", result.Nodes)
	}
	if got := result.Nodes[1].Value["statusCode"]; got != float64(200) {
		t.Errorf("statusCode = %#v, want 200", got)
	}
}

func TestGraphWorkflowWaitReceivesSignal(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	config := map[string]interface{}{
		"signal":         "approval.granted",
		"timeoutSeconds": 60,
	}
	received := WaitReceivedHandle
	timedOut := WaitTimedOutHandle
	payload := map[string]interface{}{"approvedBy": "alice"}

	// Signal before the durable timeout so the received branch wins.
	environment.RegisterDelayedCallback(func() {
		encoded, err := environment.QueryWorkflow(CurrentWaitQueryName)
		if err != nil {
			t.Errorf("QueryWorkflow(%s) error = %v", CurrentWaitQueryName, err)
			return
		}
		var wait CurrentWait
		if err := encoded.Get(&wait); err != nil {
			t.Errorf("decode current wait: %v", err)
			return
		}
		if wait.NodeID != "wait-1" || wait.Signal != "approval.granted" {
			t.Errorf("current wait = %#v, want wait-1/approval.granted", wait)
		}
		environment.SignalWorkflow("approval.granted", payload)
	}, time.Second)

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "wait-1", Type: WaitNodeType, Config: &config},
				{Id: "noop-received", Type: NoopNodeType},
				{Id: "noop-timeout", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "wait-1"},
				{Id: "e-recv", Source: "wait-1", Target: "noop-received", SourceHandle: &received},
				{Id: "e-to", Source: "wait-1", Target: "noop-timeout", SourceHandle: &timedOut},
			},
		},
	})
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "wait-1", "noop-received"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
	waitResult := result.Nodes[1].Value
	if got := waitResult["timedOut"]; got != false {
		t.Errorf("timedOut = %#v, want false", got)
	}
	if got := waitResult["branch"]; got != WaitReceivedHandle {
		t.Errorf("branch = %#v, want %q", got, WaitReceivedHandle)
	}
	gotPayload, ok := waitResult["payload"].(map[string]interface{})
	if !ok || gotPayload["approvedBy"] != "alice" {
		t.Errorf("payload = %#v, want approvedBy=alice", waitResult["payload"])
	}
}

func TestGraphWorkflowWaitTimesOut(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})

	timerFired := false
	environment.SetOnTimerFiredListener(func(string) {
		timerFired = true
	})

	config := map[string]interface{}{
		"signal":         "approval.granted",
		"timeoutSeconds": 30,
	}
	received := WaitReceivedHandle
	timedOut := WaitTimedOutHandle

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{
		Graph: api.Graph{
			Nodes: []api.Node{
				{Id: "start-1", Type: StartNodeType},
				{Id: "wait-1", Type: WaitNodeType, Config: &config},
				{Id: "noop-received", Type: NoopNodeType},
				{Id: "noop-timeout", Type: NoopNodeType},
			},
			Edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "wait-1"},
				{Id: "e-recv", Source: "wait-1", Target: "noop-received", SourceHandle: &received},
				{Id: "e-to", Source: "wait-1", Target: "noop-timeout", SourceHandle: &timedOut},
			},
		},
	})
	if !timerFired {
		t.Fatal("expected a durable timer to fire for the wait timeout")
	}
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result GraphWorkflowResult
	if err := environment.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "wait-1", "noop-timeout"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
	if got := result.Nodes[1].Value["timedOut"]; got != true {
		t.Errorf("timedOut = %#v, want true", got)
	}
	if got := result.Nodes[1].Value["branch"]; got != WaitTimedOutHandle {
		t.Errorf("branch = %#v, want %q", got, WaitTimedOutHandle)
	}
	if _, ok := result.Nodes[1].Value["payload"]; ok {
		t.Errorf("payload = %#v, want absent on timeout", result.Nodes[1].Value["payload"])
	}
}
