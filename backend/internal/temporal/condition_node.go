package temporal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/madmmas/temflowral/backend/internal/api"
)

const (
	// ConditionNodeType evaluates a predecessor value and takes the true or
	// false outgoing branch via Edge.sourceHandle. Handled in workflow code,
	// not as an activity.
	ConditionNodeType = "condition"

	// ConditionTrueHandle is the sourceHandle for the true branch.
	ConditionTrueHandle = "true"
	// ConditionFalseHandle is the sourceHandle for the false branch.
	ConditionFalseHandle = "false"

	maxConditionFieldLength = 256
)

// parseConditionNodeConfig validates a condition node's configuration.
func parseConditionNodeConfig(node api.Node) (api.ConditionNodeConfig, error) {
	if node.Config == nil {
		return api.ConditionNodeConfig{}, fmt.Errorf("condition node %q config is required", node.Id)
	}
	if _, ok := (*node.Config)["field"]; !ok {
		return api.ConditionNodeConfig{}, fmt.Errorf("condition node %q config requires \"field\"", node.Id)
	}
	if _, ok := (*node.Config)["equals"]; !ok {
		return api.ConditionNodeConfig{}, fmt.Errorf("condition node %q config requires \"equals\"", node.Id)
	}

	encoded, err := json.Marshal(*node.Config)
	if err != nil {
		return api.ConditionNodeConfig{}, fmt.Errorf("encode condition node %q config: %w", node.Id, err)
	}

	var config api.ConditionNodeConfig
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return api.ConditionNodeConfig{}, fmt.Errorf("invalid condition node %q config: %w", node.Id, err)
	}
	if config.Field == "" || len(config.Field) > maxConditionFieldLength {
		return api.ConditionNodeConfig{}, fmt.Errorf(
			"condition node %q field must be 1-%d characters",
			node.Id,
			maxConditionFieldLength,
		)
	}
	return config, nil
}

// evaluateCondition compares the first active predecessor's value[field] to
// equals using JSON equality. A missing field does not match.
func evaluateCondition(config api.ConditionNodeConfig, inputs []NodeResult) bool {
	if len(inputs) == 0 {
		return false
	}
	value, ok := inputs[0].Value[config.Field]
	if !ok {
		return false
	}
	return valuesEqual(value, config.Equals)
}

func valuesEqual(left, right interface{}) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	// Normalize through JSON so numerically equal forms (e.g. int vs float64
	// from different decode paths) compare as equal.
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	var leftNorm, rightNorm interface{}
	if err := json.Unmarshal(leftJSON, &leftNorm); err != nil {
		return false
	}
	if err := json.Unmarshal(rightJSON, &rightNorm); err != nil {
		return false
	}
	return reflect.DeepEqual(leftNorm, rightNorm)
}

func conditionHandle(matched bool) string {
	if matched {
		return ConditionTrueHandle
	}
	return ConditionFalseHandle
}

func isConditionHandle(handle string) bool {
	return handle == ConditionTrueHandle || handle == ConditionFalseHandle
}
