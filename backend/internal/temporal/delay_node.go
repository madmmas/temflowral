package temporal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// DelayNodeType pauses the workflow using a durable Temporal timer. It is
	// handled inside the workflow (not an activity) so the wait survives worker
	// restarts and does not consume an activity slot.
	DelayNodeType = "delay"

	// maxDelaySeconds bounds a single delay node to 7 days.
	maxDelaySeconds = 604800
)

// parseDelayNodeConfig validates a delay node's configuration.
func parseDelayNodeConfig(node api.Node) (api.DelayNodeConfig, error) {
	if node.Config == nil {
		return api.DelayNodeConfig{}, fmt.Errorf("delay node %q config is required", node.Id)
	}
	// Seconds is a required non-pointer field, so a missing key would silently
	// decode to 0. Enforce presence explicitly.
	if _, ok := (*node.Config)["seconds"]; !ok {
		return api.DelayNodeConfig{}, fmt.Errorf("delay node %q config requires \"seconds\"", node.Id)
	}
	encoded, err := json.Marshal(*node.Config)
	if err != nil {
		return api.DelayNodeConfig{}, fmt.Errorf("encode delay node %q config: %w", node.Id, err)
	}

	var config api.DelayNodeConfig
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return api.DelayNodeConfig{}, fmt.Errorf("invalid delay node %q config: %w", node.Id, err)
	}
	if config.Seconds < 0 || config.Seconds > maxDelaySeconds {
		return api.DelayNodeConfig{}, fmt.Errorf(
			"delay node %q seconds must be between 0 and %d",
			node.Id,
			maxDelaySeconds,
		)
	}
	return config, nil
}

// delayDuration converts a validated delay config into a timer duration.
func delayDuration(config api.DelayNodeConfig) time.Duration {
	return time.Duration(config.Seconds * float64(time.Second))
}
