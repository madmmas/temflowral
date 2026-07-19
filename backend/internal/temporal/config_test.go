package temporal

import "testing"

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{
			name: "local defaults",
			want: Config{
				Address:   defaultAddress,
				Namespace: defaultNamespace,
				TaskQueue: defaultTaskQueue,
			},
		},
		{
			name: "environment overrides",
			env: map[string]string{
				addressEnv:   "temporal.example:7233",
				namespaceEnv: "development",
				taskQueueEnv: "custom-queue",
			},
			want: Config{
				Address:   "temporal.example:7233",
				Namespace: "development",
				TaskQueue: "custom-queue",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(addressEnv, test.env[addressEnv])
			t.Setenv(namespaceEnv, test.env[namespaceEnv])
			t.Setenv(taskQueueEnv, test.env[taskQueueEnv])

			if got := ConfigFromEnv(); got != test.want {
				t.Errorf("ConfigFromEnv() = %#v, want %#v", got, test.want)
			}
		})
	}
}
