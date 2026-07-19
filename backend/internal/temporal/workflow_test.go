package temporal

import (
	"testing"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestNoopWorkflowExecutesActivity(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite
	environment := suite.NewTestWorkflowEnvironment()
	environment.RegisterActivityWithOptions(NoopActivity, activity.RegisterOptions{
		Name: NoopActivityName,
	})

	const input = "smoke-test"
	environment.ExecuteWorkflow(NoopWorkflow, input)

	if !environment.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := environment.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var output string
	if err := environment.GetWorkflowResult(&output); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if output != input {
		t.Errorf("workflow result = %q, want %q", output, input)
	}
}
