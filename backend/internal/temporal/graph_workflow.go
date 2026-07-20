package temporal

import (
	"context"
	"fmt"
	"time"

	sdktemporal "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

const (
	// GraphWorkflowName is the Temporal workflow type for graph execution.
	GraphWorkflowName = "temflowral.graph"
	// NoopNodeActivityName is the Temporal activity type for noop nodes.
	NoopNodeActivityName = "temflowral.node.noop"
)

// NodeActivityInput is the Temporal activity payload for graph nodes.
type NodeActivityInput = nodetype.ActivityInput

// NodeResult is the output of a single graph node.
type NodeResult = nodetype.Result

// GraphWorkflowInput is the payload passed when starting a graph run.
type GraphWorkflowInput struct {
	Graph api.Graph              `json:"graph"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// GraphWorkflowResult is the aggregated output of a completed graph run.
type GraphWorkflowResult struct {
	Nodes []NodeResult `json:"nodes"`
}

// GraphWorkflow walks a validated graph in topological order and dispatches
// one activity per executable node type. Multi-output nodes select a branch
// via Edge.sourceHandle; nodes on the untaken path are skipped.
func GraphWorkflow(ctx workflow.Context, input GraphWorkflowInput) (GraphWorkflowResult, error) {
	plan, err := BuildExecutionPlan(input.Graph)
	if err != nil {
		return GraphWorkflowResult{}, err
	}

	nodesByID := make(map[string]api.Node, len(input.Graph.Nodes))
	for _, node := range input.Graph.Nodes {
		nodesByID[node.Id] = node
	}

	incomingEdges := make(map[string][]api.Edge, len(input.Graph.Nodes))
	for _, edge := range input.Graph.Edges {
		incomingEdges[edge.Target] = append(incomingEdges[edge.Target], edge)
	}

	resultsByID := make(map[string]NodeResult, len(plan))
	executed := make(map[string]struct{}, len(plan))
	branchTaken := make(map[string]string)
	orderedResults := make([]NodeResult, 0, len(plan))

	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &sdktemporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	for _, node := range plan {
		inputs := activeInputs(node, incomingEdges, executed, branchTaken, resultsByID)
		if node.Type != StartNodeType && len(inputs) == 0 {
			// Not reachable via any taken edge (untaken branch).
			continue
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
		case ConditionNodeType:
			config, err := parseConditionNodeConfig(node)
			if err != nil {
				return GraphWorkflowResult{}, err
			}
			matched := evaluateCondition(config, inputs)
			handle := conditionHandle(matched)
			branchTaken[node.Id] = handle
			result = NodeResult{
				NodeID: node.Id,
				Value: map[string]interface{}{
					"type":             ConditionNodeType,
					"field":            config.Field,
					"matched":          matched,
					nodetype.BranchKey: handle,
				},
			}
		default:
			activityName, ok := CurrentRegistry().ActivityName(node.Type)
			if !ok {
				return GraphWorkflowResult{}, fmt.Errorf("unsupported node type %q on node %q", node.Type, node.Id)
			}
			err := workflow.ExecuteActivity(
				activityCtx,
				activityName,
				NodeActivityInput{Node: toActivityNode(node), Inputs: inputs},
			).Get(ctx, &result)
			if err != nil {
				return GraphWorkflowResult{}, err
			}
			if result.NodeID == "" {
				result.NodeID = node.Id
			}
			if branch, ok := nodetype.SelectedBranch(result); ok {
				branchTaken[node.Id] = branch
			}
		}

		executed[node.Id] = struct{}{}
		resultsByID[node.Id] = result
		orderedResults = append(orderedResults, result)
	}

	return GraphWorkflowResult{Nodes: orderedResults}, nil
}

func toActivityNode(node api.Node) nodetype.Node {
	return nodetype.Node{
		ID:     node.Id,
		Type:   node.Type,
		Label:  node.Label,
		Config: node.Config,
	}
}

// activeInputs returns predecessor results that reach this node on the taken
// path. Edges leaving a multi-output node are filtered by the chosen handle.
func activeInputs(
	node api.Node,
	incomingEdges map[string][]api.Edge,
	executed map[string]struct{},
	branchTaken map[string]string,
	resultsByID map[string]NodeResult,
) []NodeResult {
	edges := incomingEdges[node.Id]
	inputs := make([]NodeResult, 0, len(edges))
	for _, edge := range edges {
		if _, ok := executed[edge.Source]; !ok {
			continue
		}
		if taken, ok := branchTaken[edge.Source]; ok {
			if edge.SourceHandle == nil || *edge.SourceHandle != taken {
				continue
			}
		}
		inputs = append(inputs, resultsByID[edge.Source])
	}
	return inputs
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
		NodeID: input.Node.ID,
		Value:  value,
	}, nil
}
