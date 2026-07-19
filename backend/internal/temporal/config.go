package temporal

import "os"

const (
	addressEnv   = "TEMPORAL_ADDRESS"
	namespaceEnv = "TEMPORAL_NAMESPACE"
	taskQueueEnv = "TEMPORAL_TASK_QUEUE"

	defaultAddress   = "localhost:7233"
	defaultNamespace = "default"
	defaultTaskQueue = "temflowral"
)

// Config contains the Temporal connection and worker settings.
type Config struct {
	Address   string
	Namespace string
	TaskQueue string
}

// ConfigFromEnv returns Temporal settings with local-development defaults.
func ConfigFromEnv() Config {
	return Config{
		Address:   envOrDefault(addressEnv, defaultAddress),
		Namespace: envOrDefault(namespaceEnv, defaultNamespace),
		TaskQueue: envOrDefault(taskQueueEnv, defaultTaskQueue),
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
