package temporal

import (
	"fmt"
	"sync"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

var (
	registryMu       sync.RWMutex
	processRegistry  *nodetype.Registry
	fallbackOnce     sync.Once
	fallbackRegistry *nodetype.Registry
)

// UseRegistry installs the process-wide node-type registry used by planning,
// GraphWorkflow activity lookup, and validation. Start calls this automatically.
// Tests may call it to inject a custom registry.
func UseRegistry(registry *nodetype.Registry) {
	if registry == nil {
		panic("temporal: UseRegistry requires a non-nil registry")
	}
	_ = SwapRegistry(registry)
}

// SwapRegistry replaces the process-wide registry and returns the previous
// value (which may be nil). Intended for tests that need to restore state.
func SwapRegistry(registry *nodetype.Registry) *nodetype.Registry {
	registryMu.Lock()
	defer registryMu.Unlock()
	previous := processRegistry
	processRegistry = registry
	return previous
}

// CurrentRegistry returns the active node-type registry.
func CurrentRegistry() *nodetype.Registry {
	registryMu.RLock()
	registry := processRegistry
	registryMu.RUnlock()
	if registry != nil {
		return registry
	}

	fallbackOnce.Do(func() {
		fallbackRegistry = nodetype.NewRegistry()
		if err := RegisterBuiltins(fallbackRegistry, BuiltinOptions{}); err != nil {
			panic(fmt.Sprintf("temporal: register built-in node types: %v", err))
		}
	})
	return fallbackRegistry
}

// BuiltinOptions configures built-in activity construction at registration.
type BuiltinOptions struct {
	// HTTPAllowedHosts is passed to NewHTTPNodeActivity. Empty denies all hosts.
	HTTPAllowedHosts []string
	// HTTPActivity overrides HTTP activity construction when non-nil (tests).
	HTTPActivity *HTTPNodeActivity
}

// RegisterBuiltins registers temflowral's built-in node types on registry.
func RegisterBuiltins(registry *nodetype.Registry, options BuiltinOptions) error {
	if registry == nil {
		return fmt.Errorf("registry is required")
	}

	httpActivity := options.HTTPActivity
	if httpActivity == nil {
		var err error
		httpActivity, err = NewHTTPNodeActivity(options.HTTPAllowedHosts)
		if err != nil {
			return fmt.Errorf("configure HTTP node activity: %w", err)
		}
	}

	builtins := []nodetype.Definition{
		{
			ID:          StartNodeType,
			Name:        "Start",
			Description: "Workflow entry point",
			Category:    "core",
			Kind:        nodetype.KindWorkflow,
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"additionalProperties": false,
			},
		},
		{
			ID:           NoopNodeType,
			Name:         "No-op",
			Description:  "No-op activity used to smoke-test graph execution",
			Category:     "core",
			Kind:         nodetype.KindActivity,
			ActivityName: NoopNodeActivityName,
			Activity:     NoopNodeActivity,
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"additionalProperties": true,
			},
		},
		{
			ID:           HTTPNodeType,
			Name:         "HTTP Request",
			Description:  "Make an allowlisted outbound HTTP request",
			Category:     "integration",
			Kind:         nodetype.KindActivity,
			ActivityName: HTTPNodeActivityName,
			Activity:     httpActivity.Execute,
			ValidateConfig: func(nodeID string, config map[string]interface{}) error {
				_, err := parseHTTPNodeConfig(api.Node{Id: nodeID, Type: HTTPNodeType, Config: configPtr(config)})
				return err
			},
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"required":             []string{"method", "url"},
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"method": map[string]interface{}{
						"type": "string",
						"enum": []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
					},
					"url": map[string]interface{}{
						"type":      "string",
						"format":    "uri",
						"maxLength": 2048,
					},
					"headers": map[string]interface{}{
						"type":          "object",
						"maxProperties": 32,
						"additionalProperties": map[string]interface{}{
							"type":      "string",
							"maxLength": 8192,
						},
					},
					"body": map[string]interface{}{
						"type":      "string",
						"maxLength": 1048576,
					},
				},
			},
		},
		{
			ID:          DelayNodeType,
			Name:        "Delay",
			Description: "Pause the workflow with a durable Temporal timer",
			Category:    "core",
			Kind:        nodetype.KindWorkflow,
			ValidateConfig: func(nodeID string, config map[string]interface{}) error {
				_, err := parseDelayNodeConfig(api.Node{Id: nodeID, Type: DelayNodeType, Config: configPtr(config)})
				return err
			},
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"required":             []string{"seconds"},
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"seconds": map[string]interface{}{
						"type":    "number",
						"minimum": 0,
						"maximum": 604800,
					},
				},
			},
		},
		{
			ID:          ConditionNodeType,
			Name:        "Condition",
			Description: "Branch on a predecessor field (true/false source handles)",
			Category:    "core",
			Kind:        nodetype.KindWorkflow,
			OutputHandles: []nodetype.OutputHandle{
				{ID: ConditionTrueHandle, Label: "True"},
				{ID: ConditionFalseHandle, Label: "False"},
			},
			ValidateConfig: func(nodeID string, config map[string]interface{}) error {
				_, err := parseConditionNodeConfig(api.Node{Id: nodeID, Type: ConditionNodeType, Config: configPtr(config)})
				return err
			},
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"required":             []string{"field", "equals"},
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"field": map[string]interface{}{
						"type":      "string",
						"minLength": 1,
						"maxLength": 256,
					},
					"equals": map[string]interface{}{},
				},
			},
		},
		{
			ID:          WaitNodeType,
			Name:        "Wait for Signal",
			Description: "Suspend until a named Temporal signal arrives or a timeout elapses",
			Category:    "core",
			Kind:        nodetype.KindWorkflow,
			OutputHandles: []nodetype.OutputHandle{
				{ID: WaitReceivedHandle, Label: "Received"},
				{ID: WaitTimedOutHandle, Label: "Timed out"},
			},
			ValidateConfig: func(nodeID string, config map[string]interface{}) error {
				_, err := parseWaitNodeConfig(api.Node{Id: nodeID, Type: WaitNodeType, Config: configPtr(config)})
				return err
			},
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"required":             []string{"signal", "timeoutSeconds"},
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"signal": map[string]interface{}{
						"type":      "string",
						"minLength": 1,
						"maxLength": 128,
						"pattern":   "^[A-Za-z0-9._-]+$",
					},
					"timeoutSeconds": map[string]interface{}{
						"type":    "number",
						"minimum": 0,
						"maximum": 604800,
					},
				},
			},
		},
		{
			ID:          ChildWorkflowNodeType,
			Name:        "Child Workflow",
			Description: "Run a nested graph as a Temporal child workflow and wait for its result",
			Category:    "core",
			Kind:        nodetype.KindWorkflow,
			ValidateConfig: func(nodeID string, config map[string]interface{}) error {
				_, err := parseChildWorkflowNodeConfig(api.Node{
					Id:     nodeID,
					Type:   ChildWorkflowNodeType,
					Config: configPtr(config),
				})
				return err
			},
			ConfigSchema: map[string]interface{}{
				"type":                 "object",
				"required":             []string{"graph"},
				"additionalProperties": false,
				"properties": map[string]interface{}{
					"graph": map[string]interface{}{
						"type":                 "object",
						"required":             []string{"nodes", "edges"},
						"additionalProperties": false,
						"properties": map[string]interface{}{
							"nodes": map[string]interface{}{
								"type":     "array",
								"minItems": 1,
								"maxItems": maxChildWorkflowNodes,
							},
							"edges": map[string]interface{}{
								"type":     "array",
								"maxItems": maxChildWorkflowEdges,
							},
						},
					},
					"input": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": true,
					},
				},
			},
		},
	}

	for _, def := range builtins {
		if err := registry.Register(def); err != nil {
			return err
		}
	}
	return nil
}

func configPtr(config map[string]interface{}) *map[string]interface{} {
	if config == nil {
		return nil
	}
	return &config
}

func nodeConfigMap(node api.Node) map[string]interface{} {
	if node.Config == nil {
		return nil
	}
	return *node.Config
}
