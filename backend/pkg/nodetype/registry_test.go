package nodetype

import "testing"

func TestResolveOutputHandlesFixed(t *testing.T) {
	t.Parallel()

	ids, err := ResolveOutputHandles(Definition{
		ID: "condition",
		OutputHandles: []OutputHandle{
			{ID: "true"},
			{ID: "false"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("ResolveOutputHandles() error = %v", err)
	}
	if len(ids) != 2 || ids[0] != "true" || ids[1] != "false" {
		t.Fatalf("ids = %#v, want [true false]", ids)
	}
}

func TestResolveOutputHandlesFromConfigArrayOfObjects(t *testing.T) {
	t.Parallel()

	ids, err := ResolveOutputHandles(Definition{
		ID:                      "switch",
		OutputHandlesFromConfig: &HandlesFromConfig{Path: "branches"},
	}, map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{"id": "ok"},
			map[string]interface{}{"id": "err"},
		},
	})
	if err != nil {
		t.Fatalf("ResolveOutputHandles() error = %v", err)
	}
	if len(ids) != 2 || ids[0] != "ok" || ids[1] != "err" {
		t.Fatalf("ids = %#v, want [ok err]", ids)
	}
}

func TestResolveOutputHandlesFromConfigObjectKeys(t *testing.T) {
	t.Parallel()

	ids, err := ResolveOutputHandles(Definition{
		ID:                      "router",
		OutputHandlesFromConfig: &HandlesFromConfig{Path: "routes"},
	}, map[string]interface{}{
		"routes": map[string]interface{}{
			"a": map[string]interface{}{},
			"b": map[string]interface{}{},
		},
	})
	if err != nil {
		t.Fatalf("ResolveOutputHandles() error = %v", err)
	}
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("ids = %#v, want [a b]", ids)
	}
}

func TestRegistryRejectsDuplicatesAndInvalidKind(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	def := Definition{
		ID:           "echo",
		Name:         "Echo",
		Kind:         KindActivity,
		ConfigSchema: map[string]interface{}{"type": "object"},
		ActivityName: "example.echo",
		Activity:     func() {},
	}
	if err := registry.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := registry.Register(def); err == nil {
		t.Fatal("Register() duplicate = nil, want error")
	}
	if err := registry.Register(Definition{
		ID:           "bad",
		Name:         "Bad",
		Kind:         KindWorkflow,
		ConfigSchema: map[string]interface{}{"type": "object"},
		ActivityName: "nope",
		Activity:     func() {},
	}); err == nil {
		t.Fatal("Register() workflow with activity = nil, want error")
	}
}
