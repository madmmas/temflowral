package temporal

import (
	"context"
	"testing"

	"go.temporal.io/sdk/activity"
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

func TestGraphWorkflowDispatchesHTTPNode(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			return NodeResult{
				NodeID: input.Node.Id,
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
