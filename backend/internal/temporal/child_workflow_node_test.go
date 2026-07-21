package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestValidateChildWorkflowNodeConfig(t *testing.T) {
	t.Parallel()

	validNested := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":       "start-1",
				"type":     "start",
				"position": map[string]interface{}{"x": 0.0, "y": 0.0},
			},
			map[string]interface{}{
				"id":       "noop-1",
				"type":     "noop",
				"position": map[string]interface{}{"x": 100.0, "y": 0.0},
			},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "e1", "source": "start-1", "target": "noop-1"},
		},
	}

	tests := []struct {
		name    string
		config  *map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid",
			config: &map[string]interface{}{
				"graph": validNested,
			},
		},
		{
			name: "valid with input",
			config: &map[string]interface{}{
				"graph": validNested,
				"input": map[string]interface{}{"message": "hi"},
			},
		},
		{name: "missing config", config: nil, wantErr: true},
		{
			name:    "missing graph",
			config:  &map[string]interface{}{"input": map[string]interface{}{}},
			wantErr: true,
		},
		{
			name: "unknown property",
			config: &map[string]interface{}{
				"graph": validNested,
				"extra": true,
			},
			wantErr: true,
		},
		{
			name: "nested missing start",
			config: &map[string]interface{}{
				"graph": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{
							"id":       "noop-1",
							"type":     "noop",
							"position": map[string]interface{}{"x": 0.0, "y": 0.0},
						},
					},
					"edges": []interface{}{},
				},
			},
			wantErr: true,
		},
		{
			name: "nested childWorkflow forbidden",
			config: &map[string]interface{}{
				"graph": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{
							"id":       "start-1",
							"type":     "start",
							"position": map[string]interface{}{"x": 0.0, "y": 0.0},
						},
						map[string]interface{}{
							"id":       "child-1",
							"type":     "childWorkflow",
							"position": map[string]interface{}{"x": 100.0, "y": 0.0},
							"config": map[string]interface{}{
								"graph": validNested,
							},
						},
					},
					"edges": []interface{}{
						map[string]interface{}{"id": "e1", "source": "start-1", "target": "child-1"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := api.Node{Id: "child-1", Type: ChildWorkflowNodeType, Config: test.config}
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
