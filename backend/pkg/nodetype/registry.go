package nodetype

import (
	"fmt"
	"sort"
	"sync"
)

// Registry holds node-type definitions and their activities for discovery,
// planning, and worker startup. A single Registry instance must be shared by
// the HTTP API and the Temporal worker so ListNodeTypes matches execution.
type Registry struct {
	mu   sync.RWMutex
	defs map[string]Definition
}

// NewRegistry returns an empty node-type registry.
func NewRegistry() *Registry {
	return &Registry{defs: make(map[string]Definition)}
}

// Register adds a node-type definition. Duplicate IDs are rejected.
func (registry *Registry) Register(def Definition) error {
	if err := validateDefinition(def); err != nil {
		return err
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.defs[def.ID]; exists {
		return fmt.Errorf("node type %q is already registered", def.ID)
	}
	registry.defs[def.ID] = def
	return nil
}

// Get returns a definition by node type id.
func (registry *Registry) Get(id string) (Definition, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	def, ok := registry.defs[id]
	return def, ok
}

// List returns definitions sorted by id for stable discovery responses.
func (registry *Registry) List() []Definition {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	out := make([]Definition, 0, len(registry.defs))
	for _, def := range registry.defs {
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

// ActivityName returns the Temporal activity type for an activity-backed node.
func (registry *Registry) ActivityName(nodeType string) (string, bool) {
	def, ok := registry.Get(nodeType)
	if !ok || def.Kind != KindActivity || def.ActivityName == "" {
		return "", false
	}
	return def.ActivityName, true
}

// Activities returns activity name → implementation pairs for worker registration.
func (registry *Registry) Activities() []ActivityRegistration {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	out := make([]ActivityRegistration, 0)
	for _, def := range registry.defs {
		if def.Kind != KindActivity || def.Activity == nil || def.ActivityName == "" {
			continue
		}
		out = append(out, ActivityRegistration{
			Name: def.ActivityName,
			Fn:   def.Activity,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

// ActivityRegistration pairs a Temporal activity type name with its function.
type ActivityRegistration struct {
	Name string
	Fn   any
}

func validateDefinition(def Definition) error {
	if def.ID == "" {
		return fmt.Errorf("node type id is required")
	}
	if def.Name == "" {
		return fmt.Errorf("node type %q name is required", def.ID)
	}
	if def.ConfigSchema == nil {
		return fmt.Errorf("node type %q configSchema is required", def.ID)
	}
	switch def.Kind {
	case KindActivity:
		if def.ActivityName == "" {
			return fmt.Errorf("activity node type %q requires ActivityName", def.ID)
		}
		if def.Activity == nil {
			return fmt.Errorf("activity node type %q requires Activity", def.ID)
		}
	case KindWorkflow:
		if def.ActivityName != "" || def.Activity != nil {
			return fmt.Errorf("workflow node type %q must not set Activity or ActivityName", def.ID)
		}
	default:
		return fmt.Errorf("node type %q has invalid kind %q", def.ID, def.Kind)
	}
	if def.OutputHandlesFromConfig != nil && len(def.OutputHandles) > 0 {
		return fmt.Errorf("node type %q cannot set both outputHandles and outputHandlesFromConfig", def.ID)
	}
	if def.OutputHandlesFromConfig != nil && def.OutputHandlesFromConfig.Path == "" {
		return fmt.Errorf("node type %q outputHandlesFromConfig.path is required", def.ID)
	}
	seen := make(map[string]struct{}, len(def.OutputHandles))
	for _, handle := range def.OutputHandles {
		if handle.ID == "" {
			return fmt.Errorf("node type %q has an empty output handle id", def.ID)
		}
		if _, exists := seen[handle.ID]; exists {
			return fmt.Errorf("node type %q has duplicate output handle %q", def.ID, handle.ID)
		}
		seen[handle.ID] = struct{}{}
	}
	return nil
}
