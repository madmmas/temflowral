package temporal

import (
	"fmt"
	"slices"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

const (
	// StartNodeType is the workflow entry node. It does not run an activity.
	StartNodeType = "start"
	// NoopNodeType is the smoke executable node type.
	NoopNodeType = "noop"
)

// isExecutableNodeType reports whether the registry knows how to run a type.
func isExecutableNodeType(nodeType string) bool {
	_, ok := CurrentRegistry().Get(nodeType)
	return ok
}

// BuildExecutionPlan validates a graph and returns nodes in deterministic
// topological order for sequential Temporal activity dispatch.
func BuildExecutionPlan(graph api.Graph) ([]api.Node, error) {
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("graph has no nodes")
	}

	nodesByID := make(map[string]api.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if node.Id == "" {
			return nil, fmt.Errorf("node id is required")
		}
		if _, exists := nodesByID[node.Id]; exists {
			return nil, fmt.Errorf("duplicate node id %q", node.Id)
		}
		if node.Type == "" {
			return nil, fmt.Errorf("node %q has empty type", node.Id)
		}
		if !isExecutableNodeType(node.Type) {
			return nil, fmt.Errorf("unsupported node type %q on node %q", node.Type, node.Id)
		}
		if err := ValidateNodeConfig(node); err != nil {
			return nil, err
		}
		if err := ValidateActivityOptions(node); err != nil {
			return nil, err
		}
		nodesByID[node.Id] = node
	}

	edgeIDs := make(map[string]struct{}, len(graph.Edges))
	incoming := make(map[string][]string, len(graph.Nodes))
	outgoing := make(map[string][]string, len(graph.Nodes))
	indegree := make(map[string]int, len(graph.Nodes))
	for id := range nodesByID {
		indegree[id] = 0
	}

	for _, edge := range graph.Edges {
		if edge.Id == "" {
			return nil, fmt.Errorf("edge id is required")
		}
		if _, exists := edgeIDs[edge.Id]; exists {
			return nil, fmt.Errorf("duplicate edge id %q", edge.Id)
		}
		edgeIDs[edge.Id] = struct{}{}

		if _, ok := nodesByID[edge.Source]; !ok {
			return nil, fmt.Errorf("edge %q references missing source node %q", edge.Id, edge.Source)
		}
		if _, ok := nodesByID[edge.Target]; !ok {
			return nil, fmt.Errorf("edge %q references missing target node %q", edge.Id, edge.Target)
		}

		outgoing[edge.Source] = append(outgoing[edge.Source], edge.Target)
		incoming[edge.Target] = append(incoming[edge.Target], edge.Source)
		indegree[edge.Target]++
	}

	if err := validateOutputHandles(nodesByID, graph.Edges); err != nil {
		return nil, err
	}

	var startNodes []api.Node
	for _, node := range graph.Nodes {
		if node.Type == StartNodeType {
			startNodes = append(startNodes, node)
		}
	}
	switch len(startNodes) {
	case 0:
		return nil, fmt.Errorf("graph requires exactly one %q node", StartNodeType)
	case 1:
	default:
		return nil, fmt.Errorf("graph has %d %q nodes; want exactly one", len(startNodes), StartNodeType)
	}
	start := startNodes[0]
	if indegree[start.Id] != 0 {
		return nil, fmt.Errorf("%q node %q must not have incoming edges", StartNodeType, start.Id)
	}

	queue := []string{start.Id}
	order := make([]api.Node, 0, len(graph.Nodes))
	visited := make(map[string]struct{}, len(graph.Nodes))

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		if _, seen := visited[nodeID]; seen {
			continue
		}
		visited[nodeID] = struct{}{}
		order = append(order, nodesByID[nodeID])

		for _, targetID := range outgoing[nodeID] {
			indegree[targetID]--
			if indegree[targetID] == 0 {
				queue = append(queue, targetID)
			}
		}
	}

	if len(order) != len(graph.Nodes) {
		missing := make([]string, 0)
		for id := range nodesByID {
			if _, ok := visited[id]; !ok {
				missing = append(missing, id)
			}
		}
		slices.Sort(missing)
		return nil, fmt.Errorf("graph contains a cycle or unreachable nodes: %v", missing)
	}

	return order, nil
}

// validateOutputHandles ensures multi-output nodes expose every advertised
// handle via at least one outgoing Edge.sourceHandle.
func validateOutputHandles(nodesByID map[string]api.Node, edges []api.Edge) error {
	outgoing := make(map[string][]api.Edge)
	for _, edge := range edges {
		outgoing[edge.Source] = append(outgoing[edge.Source], edge)
	}

	for _, node := range nodesByID {
		def, ok := CurrentRegistry().Get(node.Type)
		if !ok {
			continue
		}
		handles, err := nodetype.ResolveOutputHandles(def, nodeConfigMap(node))
		if err != nil {
			return fmt.Errorf("node %q: %w", node.Id, err)
		}
		if len(handles) == 0 {
			continue
		}

		required := make(map[string]bool, len(handles))
		for _, handle := range handles {
			required[handle] = false
		}
		for _, edge := range outgoing[node.Id] {
			if edge.SourceHandle == nil {
				return fmt.Errorf(
					"node %q (type %q) edge %q requires sourceHandle matching one of %v",
					node.Id,
					node.Type,
					edge.Id,
					handles,
				)
			}
			if _, ok := required[*edge.SourceHandle]; !ok {
				return fmt.Errorf(
					"node %q (type %q) edge %q has unknown sourceHandle %q; want one of %v",
					node.Id,
					node.Type,
					edge.Id,
					*edge.SourceHandle,
					handles,
				)
			}
			required[*edge.SourceHandle] = true
		}
		for _, handle := range handles {
			if !required[handle] {
				return fmt.Errorf(
					"node %q (type %q) requires at least one outgoing edge with sourceHandle %q",
					node.Id,
					node.Type,
					handle,
				)
			}
		}
	}
	return nil
}
