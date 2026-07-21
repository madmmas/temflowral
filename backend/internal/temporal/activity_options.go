package temporal

import (
	"fmt"
	"time"

	sdktemporal "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/madmmas/temflowral/backend/internal/api"
	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

const (
	defaultActivityStartToClose = 30 * time.Second
	defaultActivityMaxAttempts  = int32(1)

	minTimeoutSeconds             = 0.001
	maxStartToCloseSeconds        = 86400.0  // 24h
	maxScheduleToCloseSeconds     = 604800.0 // 7d
	maxHeartbeatSeconds           = 3600.0
	maxInitialIntervalSeconds     = 3600.0
	maxMaximumIntervalSeconds     = 86400.0
	minBackoffCoefficient         = 1.0
	maxBackoffCoefficient         = 100.0
	maxRetryAttempts              = 100
	maxNonRetryableErrorTypeCount = 32
	maxNonRetryableErrorTypeLen   = 128
)

// defaultActivityOptions is the GraphWorkflow engine default for KindActivity
// nodes when Node.activityOptions is omitted.
func defaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: defaultActivityStartToClose,
		RetryPolicy: &sdktemporal.RetryPolicy{
			MaximumAttempts: defaultActivityMaxAttempts,
		},
	}
}

// ValidateActivityOptions ensures activityOptions is only set on activity-backed
// types and that numeric fields are within contract bounds.
func ValidateActivityOptions(node api.Node) error {
	if node.ActivityOptions == nil {
		return nil
	}
	def, ok := CurrentRegistry().Get(node.Type)
	if !ok || def.Kind != nodetype.KindActivity {
		return fmt.Errorf(
			"node %q: activityOptions is only valid for activity-backed node types",
			node.Id,
		)
	}
	_, err := activityOptionsForNode(node)
	if err != nil {
		return fmt.Errorf("node %q: %w", node.Id, err)
	}
	return nil
}

// activityOptionsForNode merges Node.activityOptions onto engine defaults.
func activityOptionsForNode(node api.Node) (workflow.ActivityOptions, error) {
	opts := defaultActivityOptions()
	if node.ActivityOptions == nil {
		return opts, nil
	}
	raw := node.ActivityOptions

	if raw.StartToCloseTimeoutSeconds != nil {
		seconds := *raw.StartToCloseTimeoutSeconds
		if err := validateTimeoutSeconds("startToCloseTimeoutSeconds", seconds, maxStartToCloseSeconds); err != nil {
			return workflow.ActivityOptions{}, err
		}
		opts.StartToCloseTimeout = secondsToDuration(seconds)
	}
	if raw.ScheduleToCloseTimeoutSeconds != nil {
		seconds := *raw.ScheduleToCloseTimeoutSeconds
		if err := validateTimeoutSeconds("scheduleToCloseTimeoutSeconds", seconds, maxScheduleToCloseSeconds); err != nil {
			return workflow.ActivityOptions{}, err
		}
		opts.ScheduleToCloseTimeout = secondsToDuration(seconds)
	}
	if raw.HeartbeatTimeoutSeconds != nil {
		seconds := *raw.HeartbeatTimeoutSeconds
		if err := validateTimeoutSeconds("heartbeatTimeoutSeconds", seconds, maxHeartbeatSeconds); err != nil {
			return workflow.ActivityOptions{}, err
		}
		opts.HeartbeatTimeout = secondsToDuration(seconds)
	}
	if raw.RetryPolicy != nil {
		policy, err := retryPolicyFromAPI(*raw.RetryPolicy)
		if err != nil {
			return workflow.ActivityOptions{}, err
		}
		opts.RetryPolicy = policy
	}
	return opts, nil
}

func retryPolicyFromAPI(raw api.RetryPolicy) (*sdktemporal.RetryPolicy, error) {
	if raw.MaximumAttempts < 1 || raw.MaximumAttempts > maxRetryAttempts {
		return nil, fmt.Errorf("retryPolicy.maximumAttempts must be between 1 and %d", maxRetryAttempts)
	}
	policy := &sdktemporal.RetryPolicy{
		MaximumAttempts: int32(raw.MaximumAttempts),
	}
	if raw.InitialIntervalSeconds != nil {
		seconds := *raw.InitialIntervalSeconds
		if err := validateTimeoutSeconds("retryPolicy.initialIntervalSeconds", seconds, maxInitialIntervalSeconds); err != nil {
			return nil, err
		}
		policy.InitialInterval = secondsToDuration(seconds)
	}
	if raw.MaximumIntervalSeconds != nil {
		seconds := *raw.MaximumIntervalSeconds
		if err := validateTimeoutSeconds("retryPolicy.maximumIntervalSeconds", seconds, maxMaximumIntervalSeconds); err != nil {
			return nil, err
		}
		policy.MaximumInterval = secondsToDuration(seconds)
	}
	if raw.BackoffCoefficient != nil {
		coefficient := *raw.BackoffCoefficient
		if coefficient < minBackoffCoefficient || coefficient > maxBackoffCoefficient {
			return nil, fmt.Errorf(
				"retryPolicy.backoffCoefficient must be between %g and %g",
				minBackoffCoefficient,
				maxBackoffCoefficient,
			)
		}
		policy.BackoffCoefficient = coefficient
	}
	if raw.NonRetryableErrorTypes != nil {
		types := *raw.NonRetryableErrorTypes
		if len(types) > maxNonRetryableErrorTypeCount {
			return nil, fmt.Errorf(
				"retryPolicy.nonRetryableErrorTypes must have at most %d items",
				maxNonRetryableErrorTypeCount,
			)
		}
		for i, name := range types {
			if name == "" || len(name) > maxNonRetryableErrorTypeLen {
				return nil, fmt.Errorf(
					"retryPolicy.nonRetryableErrorTypes[%d] must be 1-%d characters",
					i,
					maxNonRetryableErrorTypeLen,
				)
			}
		}
		policy.NonRetryableErrorTypes = append([]string(nil), types...)
	}
	return policy, nil
}

func validateTimeoutSeconds(field string, seconds, maximum float64) error {
	if seconds < minTimeoutSeconds || seconds > maximum {
		return fmt.Errorf("%s must be between %g and %g", field, minTimeoutSeconds, maximum)
	}
	return nil
}

func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
