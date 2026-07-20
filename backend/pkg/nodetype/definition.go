package nodetype

// Kind classifies how a node type is executed.
type Kind string

const (
	// KindActivity nodes dispatch a Temporal activity by ActivityName.
	KindActivity Kind = "activity"
	// KindWorkflow nodes are handled inside GraphWorkflow (timers, entry,
	// built-in branching). External registrations must use KindActivity.
	KindWorkflow Kind = "workflow"
)

// OutputHandle is a named source handle on a node type.
type OutputHandle struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

// HandlesFromConfig derives output handle IDs from a Node.config field.
// Path is a dot-separated path into config (e.g. "branches").
type HandlesFromConfig struct {
	Path string `json:"path"`
}

// ValidateConfigFunc validates a node's config map. config may be nil when
// the graph omitted Node.config. Return a clear error; do not include secrets.
type ValidateConfigFunc func(nodeID string, config map[string]interface{}) error

// Definition describes a registerable node type for discovery, planning, and
// (for KindActivity) Temporal activity dispatch.
type Definition struct {
	ID          string
	Name        string
	Description string
	Category    string
	Kind        Kind
	// ConfigSchema is JSON Schema for GET /node-types.
	ConfigSchema map[string]interface{}
	// OutputHandles is a fixed handle list. Empty means a single default
	// unnamed output. Mutually exclusive with OutputHandlesFromConfig.
	OutputHandles []OutputHandle
	// OutputHandlesFromConfig derives handles from config at plan time.
	OutputHandlesFromConfig *HandlesFromConfig
	// ValidateConfig is optional type-specific config validation.
	ValidateConfig ValidateConfigFunc
	// ActivityName is required when Kind is KindActivity.
	ActivityName string
	// Activity is the Temporal activity implementation registered on the
	// worker. Required when Kind is KindActivity. Signature should be
	// compatible with func(context.Context, ActivityInput) (Result, error)
	// or a method with that shape.
	Activity any
}
