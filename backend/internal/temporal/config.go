package temporal

import (
	"os"
	"strings"
)

const (
	addressEnv   = "TEMPORAL_ADDRESS"
	namespaceEnv = "TEMPORAL_NAMESPACE"
	taskQueueEnv = "TEMPORAL_TASK_QUEUE"
	httpHostsEnv = "HTTP_ALLOWED_HOSTS"

	defaultAddress   = "localhost:7233"
	defaultNamespace = "default"
	defaultTaskQueue = "temflowral"
)

// Config contains the Temporal connection and worker settings.
type Config struct {
	Address          string
	Namespace        string
	TaskQueue        string
	HTTPAllowedHosts []string
}

// ConfigFromEnv returns Temporal settings with local-development defaults.
func ConfigFromEnv() Config {
	return Config{
		Address:   envOrDefault(addressEnv, defaultAddress),
		Namespace: envOrDefault(namespaceEnv, defaultNamespace),
		TaskQueue: envOrDefault(taskQueueEnv, defaultTaskQueue),
		HTTPAllowedHosts: splitNonEmpty(
			os.Getenv(httpHostsEnv),
		),
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func splitNonEmpty(value string) []string {
	var values []string
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
