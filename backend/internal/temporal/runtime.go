package temporal

import (
	"context"
	"fmt"
	"sync"

	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/madmmas/temflowral/backend/pkg/nodetype"
)

// Runtime owns the process-wide Temporal client and worker.
type Runtime struct {
	client    client.Client
	worker    worker.Worker
	taskQueue string
	registry  *nodetype.Registry
	once      sync.Once
}

// WorkflowExecution identifies a started Temporal workflow.
type WorkflowExecution struct {
	ID    string
	RunID string
}

// WorkflowStatus is a Temporal workflow lifecycle snapshot for API mapping.
type WorkflowStatus struct {
	Status enums.WorkflowExecutionStatus
	Result *GraphWorkflowResult
	Error  string
}

// StartOption configures Temporal runtime startup.
type StartOption func(*startOptions)

type startOptions struct {
	registry *nodetype.Registry
}

// WithRegistry supplies a pre-built node-type registry (built-ins already
// registered, plus any external types). When omitted, Start creates a registry
// and registers built-ins using config.HTTPAllowedHosts.
func WithRegistry(registry *nodetype.Registry) StartOption {
	return func(options *startOptions) {
		options.registry = registry
	}
}

// Start connects to Temporal, registers workflows and activities, and starts
// polling the configured task queue.
func Start(config Config, opts ...StartOption) (*Runtime, error) {
	options := startOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	registry := options.registry
	if registry == nil {
		registry = nodetype.NewRegistry()
		if err := RegisterBuiltins(registry, BuiltinOptions{
			HTTPAllowedHosts: config.HTTPAllowedHosts,
		}); err != nil {
			return nil, err
		}
	}
	UseRegistry(registry)

	temporalClient, err := client.Dial(client.Options{
		HostPort:  config.Address,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("dial Temporal at %s: %w", config.Address, err)
	}

	temporalWorker := worker.New(temporalClient, config.TaskQueue, worker.Options{})
	temporalWorker.RegisterWorkflowWithOptions(NoopWorkflow, workflow.RegisterOptions{
		Name: NoopWorkflowName,
	})
	temporalWorker.RegisterActivityWithOptions(NoopActivity, activity.RegisterOptions{
		Name: NoopActivityName,
	})
	temporalWorker.RegisterWorkflowWithOptions(GraphWorkflow, workflow.RegisterOptions{
		Name: GraphWorkflowName,
	})
	for _, registration := range registry.Activities() {
		temporalWorker.RegisterActivityWithOptions(registration.Fn, activity.RegisterOptions{
			Name: registration.Name,
		})
	}

	if err := temporalWorker.Start(); err != nil {
		temporalClient.Close()
		return nil, fmt.Errorf("start Temporal worker on task queue %s: %w", config.TaskQueue, err)
	}

	return &Runtime{
		client:    temporalClient,
		worker:    temporalWorker,
		taskQueue: config.TaskQueue,
		registry:  registry,
	}, nil
}

// Registry returns the node-type registry bound to this runtime.
func (runtime *Runtime) Registry() *nodetype.Registry {
	return runtime.registry
}

// Close stops polling and closes the Temporal client. It is safe to call more
// than once.
func (runtime *Runtime) Close() {
	runtime.once.Do(func() {
		runtime.worker.Stop()
		runtime.client.Close()
	})
}

// StartGraphWorkflow starts GraphWorkflow with a stable workflow ID.
func (runtime *Runtime) StartGraphWorkflow(
	ctx context.Context,
	workflowID string,
	input GraphWorkflowInput,
) (WorkflowExecution, error) {
	run, err := runtime.client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: runtime.taskQueue,
	}, GraphWorkflowName, input)
	if err != nil {
		return WorkflowExecution{}, fmt.Errorf("start graph workflow %q: %w", workflowID, err)
	}
	return WorkflowExecution{ID: run.GetID(), RunID: run.GetRunID()}, nil
}

// DescribeGraphWorkflow reports Temporal status and, when completed, the result.
func (runtime *Runtime) DescribeGraphWorkflow(
	ctx context.Context,
	execution WorkflowExecution,
) (WorkflowStatus, error) {
	description, err := runtime.client.DescribeWorkflowExecution(ctx, execution.ID, execution.RunID)
	if err != nil {
		return WorkflowStatus{}, fmt.Errorf("describe workflow %q: %w", execution.ID, err)
	}

	status := WorkflowStatus{
		Status: description.GetWorkflowExecutionInfo().GetStatus(),
	}

	switch status.Status {
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		var result GraphWorkflowResult
		if err := runtime.client.GetWorkflow(ctx, execution.ID, execution.RunID).Get(ctx, &result); err != nil {
			return WorkflowStatus{}, fmt.Errorf("get workflow result %q: %w", execution.ID, err)
		}
		status.Result = &result
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED,
		enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
		enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		err := runtime.client.GetWorkflow(ctx, execution.ID, execution.RunID).Get(ctx, nil)
		if err != nil {
			status.Error = err.Error()
		}
	}

	return status, nil
}

// QueryCurrentWait returns which wait node GraphWorkflow is blocked on, if any.
func (runtime *Runtime) QueryCurrentWait(
	ctx context.Context,
	execution WorkflowExecution,
) (CurrentWait, error) {
	encoded, err := runtime.client.QueryWorkflow(ctx, execution.ID, execution.RunID, CurrentWaitQueryName)
	if err != nil {
		return CurrentWait{}, fmt.Errorf("query current wait on workflow %q: %w", execution.ID, err)
	}
	var wait CurrentWait
	if err := encoded.Get(&wait); err != nil {
		return CurrentWait{}, fmt.Errorf("decode current wait on workflow %q: %w", execution.ID, err)
	}
	return wait, nil
}

// SignalGraphWorkflow delivers a Temporal signal to a running graph workflow.
func (runtime *Runtime) SignalGraphWorkflow(
	ctx context.Context,
	execution WorkflowExecution,
	signalName string,
	payload interface{},
) error {
	if err := runtime.client.SignalWorkflow(ctx, execution.ID, execution.RunID, signalName, payload); err != nil {
		return fmt.Errorf("signal workflow %q (%s): %w", execution.ID, signalName, err)
	}
	return nil
}
