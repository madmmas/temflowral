package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestBuildExecutionPlanLinearGraph(t *testing.T) {
	t.Parallel()

	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "noop-1", Type: NoopNodeType},
			{Id: "noop-2", Type: NoopNodeType},
		},
		Edges: []api.Edge{
			{Id: "e1", Source: "start-1", Target: "noop-1"},
			{Id: "e2", Source: "noop-1", Target: "noop-2"},
		},
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	want := []string{"start-1", "noop-1", "noop-2"}
	if got := nodeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanPreservesEdgeOrderForFanOut(t *testing.T) {
	t.Parallel()

	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "noop-a", Type: NoopNodeType},
			{Id: "noop-b", Type: NoopNodeType},
		},
		Edges: []api.Edge{
			{Id: "e-b", Source: "start-1", Target: "noop-b"},
			{Id: "e-a", Source: "start-1", Target: "noop-a"},
		},
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	want := []string{"start-1", "noop-b", "noop-a"}
	if got := nodeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanAcceptsValidHTTPNode(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{
		"method": "GET",
		"url":    "https://api.example.com/items",
	}
	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "http-1", Type: HTTPNodeType, Config: &config},
		},
		Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "http-1"}},
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	if got, want := nodeIDs(plan), []string{"start-1", "http-1"}; !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanAcceptsValidDelayNode(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{"seconds": 5}
	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "delay-1", Type: DelayNodeType, Config: &config},
		},
		Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "delay-1"}},
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	if got, want := nodeIDs(plan), []string{"start-1", "delay-1"}; !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanRejectsInvalidDelayConfig(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{"seconds": -1}
	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "delay-1", Type: DelayNodeType, Config: &config},
		},
		Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "delay-1"}},
	}

	if _, err := BuildExecutionPlan(graph); err == nil {
		t.Fatal("BuildExecutionPlan() error = nil, want an error")
	}
}

func TestBuildExecutionPlanAcceptsValidConditionNode(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{"field": "status", "equals": "ok"}
	trueHandle := ConditionTrueHandle
	falseHandle := ConditionFalseHandle
	graph := api.Graph{
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
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	want := []string{"start-1", "cond-1", "noop-true", "noop-false"}
	if got := nodeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanAcceptsValidWaitNode(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{
		"signal":         "approval.granted",
		"timeoutSeconds": 60,
	}
	received := WaitReceivedHandle
	timedOut := WaitTimedOutHandle
	graph := api.Graph{
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
	}

	plan, err := BuildExecutionPlan(graph)
	if err != nil {
		t.Fatalf("BuildExecutionPlan() error = %v", err)
	}
	want := []string{"start-1", "wait-1", "noop-received", "noop-timeout"}
	if got := nodeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
}

func TestBuildExecutionPlanRejectsInvalidWaitBranches(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{
		"signal":         "approval.granted",
		"timeoutSeconds": 60,
	}
	received := WaitReceivedHandle

	graph := api.Graph{
		Nodes: []api.Node{
			{Id: "start-1", Type: StartNodeType},
			{Id: "wait-1", Type: WaitNodeType, Config: &config},
			{Id: "noop-received", Type: NoopNodeType},
			{Id: "noop-timeout", Type: NoopNodeType},
		},
		Edges: []api.Edge{
			{Id: "e0", Source: "start-1", Target: "wait-1"},
			{Id: "e1", Source: "wait-1", Target: "noop-received", SourceHandle: &received},
		},
	}
	if _, err := BuildExecutionPlan(graph); err == nil {
		t.Fatal("BuildExecutionPlan() error = nil, want an error")
	}
}

func TestBuildExecutionPlanRejectsInvalidConditionBranches(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{"field": "status", "equals": "ok"}
	trueHandle := ConditionTrueHandle

	tests := []struct {
		name  string
		edges []api.Edge
	}{
		{
			name: "missing sourceHandle",
			edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "cond-1"},
				{Id: "e1", Source: "cond-1", Target: "noop-true"},
				{Id: "e2", Source: "cond-1", Target: "noop-false", SourceHandle: &trueHandle},
			},
		},
		{
			name: "missing false branch",
			edges: []api.Edge{
				{Id: "e0", Source: "start-1", Target: "cond-1"},
				{Id: "e1", Source: "cond-1", Target: "noop-true", SourceHandle: &trueHandle},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			graph := api.Graph{
				Nodes: []api.Node{
					{Id: "start-1", Type: StartNodeType},
					{Id: "cond-1", Type: ConditionNodeType, Config: &config},
					{Id: "noop-true", Type: NoopNodeType},
					{Id: "noop-false", Type: NoopNodeType},
				},
				Edges: test.edges,
			}
			if _, err := BuildExecutionPlan(graph); err == nil {
				t.Fatal("BuildExecutionPlan() error = nil, want an error")
			}
		})
	}
}

func TestBuildExecutionPlanRejectsInvalidGraphs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		graph api.Graph
	}{
		{
			name: "missing start",
			graph: api.Graph{
				Nodes: []api.Node{{Id: "noop-1", Type: NoopNodeType}},
			},
		},
		{
			name: "unsupported type",
			graph: api.Graph{
				Nodes: []api.Node{
					{Id: "start-1", Type: StartNodeType},
					{Id: "unknown-1", Type: "unknown"},
				},
				Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "unknown-1"}},
			},
		},
		{
			name: "cycle",
			graph: api.Graph{
				Nodes: []api.Node{
					{Id: "start-1", Type: StartNodeType},
					{Id: "noop-1", Type: NoopNodeType},
					{Id: "noop-2", Type: NoopNodeType},
				},
				Edges: []api.Edge{
					{Id: "e1", Source: "start-1", Target: "noop-1"},
					{Id: "e2", Source: "noop-1", Target: "noop-2"},
					{Id: "e3", Source: "noop-2", Target: "noop-1"},
				},
			},
		},
		{
			name: "unreachable node",
			graph: api.Graph{
				Nodes: []api.Node{
					{Id: "start-1", Type: StartNodeType},
					{Id: "noop-1", Type: NoopNodeType},
					{Id: "orphan", Type: NoopNodeType},
				},
				Edges: []api.Edge{{Id: "e1", Source: "start-1", Target: "noop-1"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := BuildExecutionPlan(test.graph); err == nil {
				t.Fatal("BuildExecutionPlan() error = nil, want an error")
			}
		})
	}
}

func nodeIDs(nodes []api.Node) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.Id)
	}
	return ids
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
