package temporal

import (
	"context"
	"fmt"
	"time"

	sdktemporal "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// GraphWorkflowName is the Temporal workflow type for graph execution.
	GraphWorkflowName = "temflowral.graph"
	// NoopNodeActivityName is the Temporal activity type for noop nodes.
	NoopNodeActivityName = "temflowral.node.noop"
)

var activityByNodeType = map[string]string{
	NoopNodeType: NoopNodeActivityName,
	HTTPNodeType: HTTPNodeActivityName,
}

// GraphWorkflowInput is the payload passed when starting a graph run.
type GraphWorkflowInput struct {
	Graph api.Graph              `json:"graph"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// NodeActivityInput is passed to node activities.
type NodeActivityInput struct {
	Node   api.Node     `json:"node"`
	Inputs []NodeResult `json:"inputs"`
}

// NodeResult is the output of a single graph node.
type NodeResult struct {
	NodeID string                 `json:"nodeId"`
	Value  map[string]interface{} `json:"value,omitempty"`
}

// GraphWorkflowResult is the aggregated output of a completed graph run.
type GraphWorkflowResult struct {
	Nodes []NodeResult `json:"nodes"`
}

// GraphWorkflow walks a validated graph in topological order and dispatches
// one activity per executable node type.
func GraphWorkflow(ctx workflow.Context, input GraphWorkflowInput) (GraphWorkflowResult, error) {
	plan, err := BuildExecutionPlan(input.Graph)
	if err != nil {
		return GraphWorkflowResult{}, err
	}

	incoming := make(map[string][]string, len(input.Graph.Nodes))
	for _, edge := range input.Graph.Edges {
		incoming[edge.Target] = append(incoming[edge.Target], edge.Source)
	}

	resultsByID := make(map[string]NodeResult, len(plan))
	orderedResults := make([]NodeResult, 0, len(plan))

	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &sdktemporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	for _, node := range plan {
		predecessorIDs := incoming[node.Id]
		inputs := make([]NodeResult, 0, len(predecessorIDs))
		for _, predecessorID := range predecessorIDs {
			inputs = append(inputs, resultsByID[predecessorID])
		}

		var result NodeResult
		switch node.Type {
		case StartNodeType:
			value := map[string]interface{}{}
			if input.Input != nil {
				value = input.Input
			}
			result = NodeResult{NodeID: node.Id, Value: value}
		case DelayNodeType:
			config, err := parseDelayNodeConfig(node)
			if err != nil {
				return GraphWorkflowResult{}, err
			}
			// Durable timer: runs in workflow code so it survives worker
			// restarts instead of blocking an activity worker.
			if err := workflow.Sleep(ctx, delayDuration(config)); err != nil {
				return GraphWorkflowResult{}, err
			}
			result = NodeResult{
				NodeID: node.Id,
				Value: map[string]interface{}{
					"type":    DelayNodeType,
					"seconds": config.Seconds,
				},
			}
		default:
			activityName, ok := activityByNodeType[node.Type]
			if !ok {
				return GraphWorkflowResult{}, fmt.Errorf("unsupported node type %q on node %q", node.Type, node.Id)
			}
			err := workflow.ExecuteActivity(
				activityCtx,
				activityName,
				NodeActivityInput{Node: node, Inputs: inputs},
			).Get(ctx, &result)
			if err != nil {
				return GraphWorkflowResult{}, err
			}
			if result.NodeID == "" {
				result.NodeID = node.Id
			}
		}

		resultsByID[node.Id] = result
		orderedResults = append(orderedResults, result)
	}

	return GraphWorkflowResult{Nodes: orderedResults}, nil
}

// NoopNodeActivity returns a trivial success payload for smoke graph runs.
func NoopNodeActivity(_ context.Context, input NodeActivityInput) (NodeResult, error) {
	value := map[string]interface{}{
		"type": NoopNodeType,
	}
	if input.Node.Config != nil {
		value["config"] = *input.Node.Config
	}
	if len(input.Inputs) > 0 {
		value["inputs"] = input.Inputs
	}
	return NodeResult{
		NodeID: input.Node.Id,
		Value:  value,
	}, nil
}
