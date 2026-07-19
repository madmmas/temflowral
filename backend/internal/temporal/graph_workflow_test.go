package temporal

import (
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
