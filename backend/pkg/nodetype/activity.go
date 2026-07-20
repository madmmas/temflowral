package nodetype

// ActivityInput is passed to registered node activities. External packages
// implement activities with this shape so they do not depend on internal APIs.
type ActivityInput struct {
	Node   Node     `json:"node"`
	Inputs []Result `json:"inputs"`
}

// Node is the activity-facing graph node payload (id, type, config).
type Node struct {
	ID     string                  `json:"id"`
	Type   string                  `json:"type"`
	Label  *string                 `json:"label,omitempty"`
	Config *map[string]interface{} `json:"config,omitempty"`
}

// Result is the output of a single graph node. Multi-output activity nodes
// select a branch by setting Value["branch"] to a handle ID advertised by
// their Definition (fixed or config-derived).
type Result struct {
	NodeID string                 `json:"nodeId"`
	Value  map[string]interface{} `json:"value,omitempty"`
}

// BranchKey is the Value map key activities use to report the taken output
// handle for multi-output nodes.
const BranchKey = "branch"

// SelectedBranch returns the handle ID from result.Value["branch"] when set.
func SelectedBranch(result Result) (string, bool) {
	if result.Value == nil {
		return "", false
	}
	raw, ok := result.Value[BranchKey]
	if !ok {
		return "", false
	}
	branch, ok := raw.(string)
	if !ok || branch == "" {
		return "", false
	}
	return branch, true
}
