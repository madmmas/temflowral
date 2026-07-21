package temporal

import (
	"testing"
	"time"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func floatPtr(v float64) *float64 { return &v }

func TestValidateActivityOptionsRejectsWorkflowNative(t *testing.T) {
	t.Parallel()

	options := &api.ActivityOptions{
		StartToCloseTimeoutSeconds: floatPtr(60),
	}
	for _, nodeType := range []string{StartNodeType, DelayNodeType, ConditionNodeType, WaitNodeType} {
		node := api.Node{Id: "n1", Type: nodeType, ActivityOptions: options}
		if err := ValidateActivityOptions(node); err == nil {
			t.Fatalf("ValidateActivityOptions(%s) error = nil, want an error", nodeType)
		}
	}
}

func TestValidateTaskQueueRejectsWorkflowNative(t *testing.T) {
	t.Parallel()

	queue := "special.queue"
	for _, nodeType := range []string{StartNodeType, DelayNodeType, ConditionNodeType, WaitNodeType} {
		node := api.Node{Id: "n1", Type: nodeType, TaskQueue: &queue}
		if err := ValidateActivityOptions(node); err == nil {
			t.Fatalf("ValidateActivityOptions(%s taskQueue) error = nil, want an error", nodeType)
		}
	}
}

func TestValidateActivityOptionsAcceptsActivityNodes(t *testing.T) {
	t.Parallel()

	options := &api.ActivityOptions{
		StartToCloseTimeoutSeconds: floatPtr(60),
		RetryPolicy: &api.RetryPolicy{
			MaximumAttempts:        3,
			InitialIntervalSeconds: floatPtr(1),
			BackoffCoefficient:     floatPtr(2),
		},
	}
	queue := "temflowral.http"
	for _, nodeType := range []string{NoopNodeType, HTTPNodeType} {
		node := api.Node{Id: "n1", Type: nodeType, ActivityOptions: options, TaskQueue: &queue}
		if err := ValidateActivityOptions(node); err != nil {
			t.Fatalf("ValidateActivityOptions(%s) error = %v", nodeType, err)
		}
	}
}

func TestActivityOptionsForNodeDefaultsAndMerge(t *testing.T) {
	t.Parallel()

	defaults, err := activityOptionsForNode(api.Node{Id: "n1", Type: NoopNodeType})
	if err != nil {
		t.Fatalf("defaults error = %v", err)
	}
	if defaults.StartToCloseTimeout != 30*time.Second {
		t.Fatalf("default StartToCloseTimeout = %v, want 30s", defaults.StartToCloseTimeout)
	}
	if defaults.RetryPolicy == nil || defaults.RetryPolicy.MaximumAttempts != 1 {
		t.Fatalf("default RetryPolicy = %#v, want MaximumAttempts 1", defaults.RetryPolicy)
	}
	if defaults.TaskQueue != "" {
		t.Fatalf("default TaskQueue = %q, want empty", defaults.TaskQueue)
	}

	queue := "worker.gpu"
	merged, err := activityOptionsForNode(api.Node{
		Id:   "n1",
		Type: NoopNodeType,
		ActivityOptions: &api.ActivityOptions{
			StartToCloseTimeoutSeconds: floatPtr(120),
		},
		TaskQueue: &queue,
	})
	if err != nil {
		t.Fatalf("merge error = %v", err)
	}
	if merged.StartToCloseTimeout != 120*time.Second {
		t.Fatalf("merged StartToCloseTimeout = %v, want 120s", merged.StartToCloseTimeout)
	}
	if merged.RetryPolicy == nil || merged.RetryPolicy.MaximumAttempts != 1 {
		t.Fatalf("merged RetryPolicy = %#v, want MaximumAttempts 1", merged.RetryPolicy)
	}
	if merged.TaskQueue != queue {
		t.Fatalf("merged TaskQueue = %q, want %q", merged.TaskQueue, queue)
	}
}

func TestActivityOptionsForNodeRejectsInvalidTaskQueue(t *testing.T) {
	t.Parallel()

	invalid := "bad queue"
	_, err := activityOptionsForNode(api.Node{
		Id:        "n1",
		Type:      NoopNodeType,
		TaskQueue: &invalid,
	})
	if err == nil {
		t.Fatal("activityOptionsForNode() error = nil, want an error")
	}
}

func TestActivityOptionsForNodeRejectsInvalidBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options *api.ActivityOptions
	}{
		{
			name: "timeout too small",
			options: &api.ActivityOptions{
				StartToCloseTimeoutSeconds: floatPtr(0),
			},
		},
		{
			name: "timeout too large",
			options: &api.ActivityOptions{
				StartToCloseTimeoutSeconds: floatPtr(maxStartToCloseSeconds + 1),
			},
		},
		{
			name: "maximumAttempts zero",
			options: &api.ActivityOptions{
				RetryPolicy: &api.RetryPolicy{MaximumAttempts: 0},
			},
		},
		{
			name: "maximumAttempts too large",
			options: &api.ActivityOptions{
				RetryPolicy: &api.RetryPolicy{MaximumAttempts: maxRetryAttempts + 1},
			},
		},
		{
			name: "backoff too small",
			options: &api.ActivityOptions{
				RetryPolicy: &api.RetryPolicy{
					MaximumAttempts:    2,
					BackoffCoefficient: floatPtr(0.5),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := activityOptionsForNode(api.Node{
				Id:              "n1",
				Type:            NoopNodeType,
				ActivityOptions: test.options,
			})
			if err == nil {
				t.Fatal("activityOptionsForNode() error = nil, want an error")
			}
		})
	}
}
