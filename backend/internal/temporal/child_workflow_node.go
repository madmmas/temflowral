package temporal

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// ChildWorkflowNodeType runs a nested graph as a Temporal child workflow.
	ChildWorkflowNodeType = "childWorkflow"

	maxChildWorkflowNodes = 50
	maxChildWorkflowEdges = 100
)

// parseChildWorkflowNodeConfig validates a childWorkflow node's configuration.
func parseChildWorkflowNodeConfig(node api.Node) (api.ChildWorkflowNodeConfig, error) {
	if node.Config == nil {
		return api.ChildWorkflowNodeConfig{}, fmt.Errorf("childWorkflow node %q config is required", node.Id)
	}
	if _, ok := (*node.Config)["graph"]; !ok {
		return api.ChildWorkflowNodeConfig{}, fmt.Errorf("childWorkflow node %q config requires \"graph\"", node.Id)
	}

	encoded, err := json.Marshal(*node.Config)
	if err != nil {
		return api.ChildWorkflowNodeConfig{}, fmt.Errorf("encode childWorkflow node %q config: %w", node.Id, err)
	}

	var config api.ChildWorkflowNodeConfig
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return api.ChildWorkflowNodeConfig{}, fmt.Errorf("invalid childWorkflow node %q config: %w", node.Id, err)
	}
	if err := validateNestedGraph(node.Id, config.Graph); err != nil {
		return api.ChildWorkflowNodeConfig{}, err
	}
	return config, nil
}

func validateNestedGraph(parentNodeID string, nested api.NestedGraph) error {
	if len(nested.Nodes) == 0 {
		return fmt.Errorf("childWorkflow node %q graph.nodes must not be empty", parentNodeID)
	}
	if len(nested.Nodes) > maxChildWorkflowNodes {
		return fmt.Errorf(
			"childWorkflow node %q graph.nodes must have at most %d items",
			parentNodeID,
			maxChildWorkflowNodes,
		)
	}
	if len(nested.Edges) > maxChildWorkflowEdges {
		return fmt.Errorf(
			"childWorkflow node %q graph.edges must have at most %d items",
			parentNodeID,
			maxChildWorkflowEdges,
		)
	}
	for _, child := range nested.Nodes {
		if child.Type == ChildWorkflowNodeType {
			return fmt.Errorf(
				"childWorkflow node %q graph must not contain nested childWorkflow nodes (found %q)",
				parentNodeID,
				child.Id,
			)
		}
	}

	graph := toRunnableGraph(nested)
	if _, err := BuildExecutionPlan(graph); err != nil {
		return fmt.Errorf("childWorkflow node %q nested graph: %w", parentNodeID, err)
	}
	return nil
}

func toRunnableGraph(nested api.NestedGraph) api.Graph {
	return api.Graph{
		Nodes: nested.Nodes,
		Edges: nested.Edges,
	}
}

func resolveChildWorkflowInput(
	config api.ChildWorkflowNodeConfig,
	inputs []NodeResult,
) map[string]interface{} {
	if config.Input != nil {
		return *config.Input
	}
	if len(inputs) > 0 && inputs[0].Value != nil {
		return inputs[0].Value
	}
	return map[string]interface{}{}
}

func childWorkflowResultValue(childResult GraphWorkflowResult) map[string]interface{} {
	nodes := make([]map[string]interface{}, 0, len(childResult.Nodes))
	for _, node := range childResult.Nodes {
		nodes = append(nodes, map[string]interface{}{
			"nodeId": node.NodeID,
			"value":  node.Value,
		})
	}
	return map[string]interface{}{
		"type":  ChildWorkflowNodeType,
		"nodes": nodes,
	}
}
