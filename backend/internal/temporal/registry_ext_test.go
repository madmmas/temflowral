package temporal

import (
	"context"
	"testing"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

func TestExternalNodeTypeRegistration(t *testing.T) {
	// Serial: mutates the process-wide registry used by GraphWorkflow.
	registry := nodetype.NewRegistry()
	if err := RegisterBuiltins(registry, BuiltinOptions{}); err != nil {
		t.Fatalf("RegisterBuiltins() error = %v", err)
	}

	const (
		externalType         = "example.switch"
		externalActivityName = "example.activity.switch"
	)
	if err := registry.Register(nodetype.Definition{
		ID:          externalType,
		Name:        "Switch",
		Description: "Example external multi-output activity",
		Category:    "example",
		Kind:        nodetype.KindActivity,
		ConfigSchema: map[string]interface{}{
			"type":                 "object",
			"required":             []string{"branches"},
			"additionalProperties": false,
			"properties": map[string]interface{}{
				"branches": map[string]interface{}{
					"type": "object",
					"additionalProperties": map[string]interface{}{
						"type": "object",
					},
				},
			},
		},
		OutputHandlesFromConfig: &nodetype.HandlesFromConfig{Path: "branches"},
		ActivityName:            externalActivityName,
		Activity: func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			return NodeResult{
				NodeID: input.Node.ID,
				Value: map[string]interface{}{
					"type":             externalType,
					nodetype.BranchKey: "left",
				},
			}, nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	previous := SwapRegistry(registry)
	t.Cleanup(func() { SwapRegistry(previous) })
	trueHandle := "left"
	falseHandle := "right"
	config := map[string]interface{}{
		"branches": map[string]interface{}{
			"left":  map[string]interface{}{},
			"right": map[string]interface{}{},
		},
	}
	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "switch-1", Type: externalType, Config: &config},
			{Id: "noop-left", Type: NoopNodeType},
			{Id: "noop-right", Type: NoopNodeType},
		},
		Edges: []api.Edge{
			{Id: "e1", Source: "start-1", Target: "switch-1"},
			{Id: "e2", Source: "switch-1", Target: "noop-left", SourceHandle: &trueHandle},
			{Id: "e3", Source: "switch-1", Target: "noop-right", SourceHandle: &falseHandle},
		},
	}

	if _, err := BuildExecutionPlan(graph); err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopNodeActivity, activity.RegisterOptions{
		Name: NoopNodeActivityName,
	})
	environment.RegisterActivityWithOptions(
		func(_ context.Context, input NodeActivityInput) (NodeResult, error) {
			return NodeResult{
				NodeID: input.Node.ID,
				Value: map[string]interface{}{
					"type":             externalType,
					nodetype.BranchKey: "left",
				},
			}, nil
		},
		activity.RegisterOptions{Name: externalActivityName},
	)

	environment.ExecuteWorkflow(GraphWorkflow, GraphWorkflowInput{Graph: graph})
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
	got := make([]string, 0, len(result.Nodes))
	for _, node := range result.Nodes {
		got = append(got, node.NodeID)
	}
	want := []string{"start-1", "switch-1", "noop-left"}
	if !equalStrings(got, want) {
		t.Fatalf("executed nodes = %v, want %v", got, want)
	}
}
