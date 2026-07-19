package temporal

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	// NoopWorkflowName is the stable workflow type used by the local smoke test.
	NoopWorkflowName = "temflowral.noop"
	// NoopActivityName is the stable activity type called by NoopWorkflow.
	NoopActivityName = "temflowral.noop.activity"
)

// NoopWorkflow executes one activity and returns its input unchanged.
func NoopWorkflow(ctx workflow.Context, input string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var output string
	err := workflow.ExecuteActivity(ctx, NoopActivityName, input).Get(ctx, &output)
	return output, err
}

// NoopActivity returns its input unchanged.
func NoopActivity(_ context.Context, input string) (string, error) {
	return input, nil
}
