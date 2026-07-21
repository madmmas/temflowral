package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestResolveTemplateStringExact(t *testing.T) {
	t.Parallel()

	root := templateDataRoot([]NodeResult{
		{NodeID: "start-1", Value: map[string]interface{}{"id": "abc", "count": float64(3)}},
	})

	got, err := resolveTemplateString("{{ nodes.start-1.output.id }}", root, newTemplateBudget())
	if err != nil {
		t.Fatalf("resolve exact string: %v", err)
	}
	if got != "abc" {
		t.Fatalf("got %#v, want abc", got)
	}

	got, err = resolveTemplateString("{{ nodes.start-1.output.count }}", root, newTemplateBudget())
	if err != nil {
		t.Fatalf("resolve exact number: %v", err)
	}
	if got != float64(3) {
		t.Fatalf("got %#v, want 3", got)
	}
}

func TestResolveTemplateStringInterpolation(t *testing.T) {
	t.Parallel()

	root := templateDataRoot([]NodeResult{
		{NodeID: "start-1", Value: map[string]interface{}{"id": "abc"}},
	})
	got, err := resolveTemplateString(
		"https://api.example.com/items/{{ nodes.start-1.output.id }}",
		root,
		newTemplateBudget(),
	)
	if err != nil {
		t.Fatalf("resolve interpolation: %v", err)
	}
	if got != "https://api.example.com/items/abc" {
		t.Fatalf("got %#v", got)
	}
}

func TestResolveTemplateErrors(t *testing.T) {
	t.Parallel()

	root := templateDataRoot([]NodeResult{
		{NodeID: "start-1", Value: map[string]interface{}{
			"id":     "abc",
			"nested": map[string]interface{}{"a": 1},
		}},
	})

	tests := []struct {
		name string
		raw  string
	}{
		{name: "unknown node", raw: "{{ nodes.missing.output.id }}"},
		{name: "missing field", raw: "{{ nodes.start-1.output.nope }}"},
		{name: "bad path", raw: "{{ env.SECRET }}"},
		{name: "filter", raw: "{{ nodes.start-1.output.id | upper }}"},
		{name: "object in interpolation", raw: "x-{{ nodes.start-1.output.nested }}-y"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := resolveTemplateString(test.raw, root, newTemplateBudget()); err == nil {
				t.Fatal("error = nil, want an error")
			}
		})
	}
}

func TestNodeWithResolvedConfigSkipsGraph(t *testing.T) {
	t.Parallel()

	nestedURL := "{{ nodes.outer.output.url }}"
	config := map[string]interface{}{
		"graph": map[string]interface{}{
			"nodes": []interface{}{
				map[string]interface{}{
					"id": "http-1",
					"config": map[string]interface{}{
						"url": nestedURL,
					},
				},
			},
		},
		"input": map[string]interface{}{
			"message": "{{ nodes.start-1.output.message }}",
		},
	}
	node := api.Node{Id: "child-1", Type: ChildWorkflowNodeType, Config: &config}
	resolved, err := nodeWithResolvedConfig(node, []NodeResult{
		{NodeID: "start-1", Value: map[string]interface{}{"message": "hi"}},
		{NodeID: "outer", Value: map[string]interface{}{"url": "https://evil.example"}},
	})
	if err != nil {
		t.Fatalf("nodeWithResolvedConfig: %v", err)
	}
	input := (*resolved.Config)["input"].(map[string]interface{})
	if input["message"] != "hi" {
		t.Fatalf("input.message = %#v, want hi", input["message"])
	}
	graph := (*resolved.Config)["graph"].(map[string]interface{})
	nodes := graph["nodes"].([]interface{})
	nestedConfig := nodes[0].(map[string]interface{})["config"].(map[string]interface{})
	if nestedConfig["url"] != nestedURL {
		t.Fatalf("nested url = %#v, want unresolved template", nestedConfig["url"])
	}
}

func TestNodeWithResolvedConfigRejectsWaitTemplates(t *testing.T) {
	t.Parallel()

	config := map[string]interface{}{
		"signal":         "{{ nodes.start-1.output.signal }}",
		"timeoutSeconds": 1.0,
	}
	node := api.Node{Id: "wait-1", Type: WaitNodeType, Config: &config}
	_, err := nodeWithResolvedConfig(node, []NodeResult{
		{NodeID: "start-1", Value: map[string]interface{}{"signal": "ready"}},
	})
	if err == nil {
		t.Fatal("error = nil, want wait template rejection")
	}
}
