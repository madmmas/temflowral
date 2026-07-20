package nodetype

import (
	"fmt"
	"sort"
	"strings"
)

// ResolveOutputHandles returns the output handle IDs for a node instance.
// Fixed handles ignore config. Config-derived handles read Definition's path.
// An empty result means a single unnamed default output (no sourceHandle).
func ResolveOutputHandles(def Definition, config map[string]interface{}) ([]string, error) {
	if def.OutputHandlesFromConfig != nil && len(def.OutputHandles) > 0 {
		return nil, fmt.Errorf(
			"node type %q cannot set both outputHandles and outputHandlesFromConfig",
			def.ID,
		)
	}
	if def.OutputHandlesFromConfig != nil {
		return handlesFromConfig(def.ID, def.OutputHandlesFromConfig.Path, config)
	}
	if len(def.OutputHandles) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(def.OutputHandles))
	seen := make(map[string]struct{}, len(def.OutputHandles))
	for _, handle := range def.OutputHandles {
		if handle.ID == "" {
			return nil, fmt.Errorf("node type %q has an empty output handle id", def.ID)
		}
		if _, exists := seen[handle.ID]; exists {
			return nil, fmt.Errorf("node type %q has duplicate output handle %q", def.ID, handle.ID)
		}
		seen[handle.ID] = struct{}{}
		ids = append(ids, handle.ID)
	}
	return ids, nil
}

func handlesFromConfig(typeID, path string, config map[string]interface{}) ([]string, error) {
	if path == "" {
		return nil, fmt.Errorf("node type %q outputHandlesFromConfig.path is required", typeID)
	}
	if config == nil {
		return nil, fmt.Errorf("node type %q requires config to derive output handles from %q", typeID, path)
	}
	value, err := lookupPath(config, path)
	if err != nil {
		return nil, fmt.Errorf("node type %q config path %q: %w", typeID, path, err)
	}
	ids, err := handleIDsFromValue(value)
	if err != nil {
		return nil, fmt.Errorf("node type %q config path %q: %w", typeID, path, err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("node type %q config path %q produced no output handles", typeID, path)
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			return nil, fmt.Errorf("node type %q config path %q contains an empty handle id", typeID, path)
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("node type %q config path %q has duplicate handle %q", typeID, path, id)
		}
		seen[id] = struct{}{}
	}
	return ids, nil
}

func lookupPath(root map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	var current interface{} = root
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty path segment")
		}
		object, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot traverse %q on non-object", part)
		}
		next, ok := object[part]
		if !ok {
			return nil, fmt.Errorf("missing key %q", part)
		}
		current = next
	}
	return current, nil
}

func handleIDsFromValue(value interface{}) ([]string, error) {
	switch typed := value.(type) {
	case []string:
		out := make([]string, len(typed))
		copy(out, typed)
		return out, nil
	case []interface{}:
		out := make([]string, 0, len(typed))
		for i, item := range typed {
			switch entry := item.(type) {
			case string:
				out = append(out, entry)
			case map[string]interface{}:
				raw, ok := entry["id"]
				if !ok {
					return nil, fmt.Errorf("array item %d missing \"id\"", i)
				}
				id, ok := raw.(string)
				if !ok {
					return nil, fmt.Errorf("array item %d \"id\" must be a string", i)
				}
				out = append(out, id)
			default:
				return nil, fmt.Errorf("array item %d must be a string or object with \"id\"", i)
			}
		}
		return out, nil
	case map[string]interface{}:
		out := make([]string, 0, len(typed))
		for key := range typed {
			out = append(out, key)
		}
		sort.Strings(out)
		return out, nil
	default:
		return nil, fmt.Errorf("want string array, object-array with id, or object; got %T", value)
	}
}
