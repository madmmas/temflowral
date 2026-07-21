package temporal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// WaitNodeType suspends the workflow until a named Temporal signal arrives
	// or a durable timeout elapses. Handled in workflow code, not as an activity.
	WaitNodeType = "wait"

	// WaitReceivedHandle is the sourceHandle when the signal arrives first.
	WaitReceivedHandle = "received"
	// WaitTimedOutHandle is the sourceHandle when the timeout wins.
	WaitTimedOutHandle = "timedOut"

	maxWaitSignalLength = 128
	maxWaitSeconds      = 604800 // 7 days, same cap as delay
)

var waitSignalPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// parseWaitNodeConfig validates a wait node's configuration.
func parseWaitNodeConfig(node api.Node) (api.WaitNodeConfig, error) {
	if node.Config == nil {
		return api.WaitNodeConfig{}, fmt.Errorf("wait node %q config is required", node.Id)
	}
	if _, ok := (*node.Config)["signal"]; !ok {
		return api.WaitNodeConfig{}, fmt.Errorf("wait node %q config requires \"signal\"", node.Id)
	}
	if _, ok := (*node.Config)["timeoutSeconds"]; !ok {
		return api.WaitNodeConfig{}, fmt.Errorf("wait node %q config requires \"timeoutSeconds\"", node.Id)
	}

	encoded, err := json.Marshal(*node.Config)
	if err != nil {
		return api.WaitNodeConfig{}, fmt.Errorf("encode wait node %q config: %w", node.Id, err)
	}

	var config api.WaitNodeConfig
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return api.WaitNodeConfig{}, fmt.Errorf("invalid wait node %q config: %w", node.Id, err)
	}
	if config.Signal == "" || utf8.RuneCountInString(config.Signal) > maxWaitSignalLength {
		return api.WaitNodeConfig{}, fmt.Errorf(
			"wait node %q signal must be 1-%d characters",
			node.Id,
			maxWaitSignalLength,
		)
	}
	if !waitSignalPattern.MatchString(config.Signal) {
		return api.WaitNodeConfig{}, fmt.Errorf(
			"wait node %q signal must match %s",
			node.Id,
			waitSignalPattern.String(),
		)
	}
	if config.TimeoutSeconds < 0 || config.TimeoutSeconds > maxWaitSeconds {
		return api.WaitNodeConfig{}, fmt.Errorf(
			"wait node %q timeoutSeconds must be between 0 and %d",
			node.Id,
			maxWaitSeconds,
		)
	}
	return config, nil
}

func waitTimeoutDuration(config api.WaitNodeConfig) time.Duration {
	return time.Duration(config.TimeoutSeconds * float64(time.Second))
}

func waitHandle(timedOut bool) string {
	if timedOut {
		return WaitTimedOutHandle
	}
	return WaitReceivedHandle
}
