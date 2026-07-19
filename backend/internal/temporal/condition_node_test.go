package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestValidateConditionNodeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *map[string]interface{}
		wantErr bool
	}{
		{name: "valid string", config: &map[string]interface{}{"field": "status", "equals": "ok"}},
		{name: "valid number", config: &map[string]interface{}{"field": "count", "equals": 3.0}},
		{name: "valid bool", config: &map[string]interface{}{"field": "ready", "equals": true}},
		{name: "missing config", config: nil, wantErr: true},
		{name: "missing field", config: &map[string]interface{}{"equals": "ok"}, wantErr: true},
		{name: "missing equals", config: &map[string]interface{}{"field": "status"}, wantErr: true},
		{name: "empty field", config: &map[string]interface{}{"field": "", "equals": "ok"}, wantErr: true},
		{name: "unknown property", config: &map[string]interface{}{"field": "status", "equals": "ok", "expr": "x"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := api.Node{Id: "cond-1", Type: ConditionNodeType, Config: test.config}
			err := ValidateNodeConfig(node)
			if test.wantErr && err == nil {
				t.Fatal("ValidateNodeConfig() error = nil, want an error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateNodeConfig() error = %v", err)
			}
		})
	}
}

func TestEvaluateCondition(t *testing.T) {
	t.Parallel()

	config := api.ConditionNodeConfig{Field: "status", Equals: "ok"}
	if !evaluateCondition(config, []NodeResult{{Value: map[string]interface{}{"status": "ok"}}}) {
		t.Fatal("expected match")
	}
	if evaluateCondition(config, []NodeResult{{Value: map[string]interface{}{"status": "fail"}}}) {
		t.Fatal("expected non-match")
	}
	if evaluateCondition(config, []NodeResult{{Value: map[string]interface{}{}}}) {
		t.Fatal("missing field should not match")
	}
	if evaluateCondition(config, nil) {
		t.Fatal("empty inputs should not match")
	}
}
