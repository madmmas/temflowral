package temporal

import (
	"fmt"
	"sync"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Runtime owns the process-wide Temporal client and worker.
type Runtime struct {
	client client.Client
	worker worker.Worker
	once   sync.Once
}

// Start connects to Temporal, registers workflows and activities, and starts
// polling the configured task queue.
func Start(config Config) (*Runtime, error) {
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

	if err := temporalWorker.Start(); err != nil {
		temporalClient.Close()
		return nil, fmt.Errorf("start Temporal worker on task queue %s: %w", config.TaskQueue, err)
	}

	return &Runtime{
		client: temporalClient,
		worker: temporalWorker,
	}, nil
}

// Close stops polling and closes the Temporal client. It is safe to call more
// than once.
func (runtime *Runtime) Close() {
	runtime.once.Do(func() {
		runtime.worker.Stop()
		runtime.client.Close()
	})
}
