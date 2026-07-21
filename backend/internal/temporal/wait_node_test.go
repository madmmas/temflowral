package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestValidateWaitNodeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid",
			config: &map[string]interface{}{
				"signal":         "approval.granted",
				"timeoutSeconds": 60,
			},
		},
		{
			name: "zero timeout",
			config: &map[string]interface{}{
				"signal":         "ready",
				"timeoutSeconds": 0,
			},
		},
		{name: "missing config", config: nil, wantErr: true},
		{
			name:    "missing signal",
			config:  &map[string]interface{}{"timeoutSeconds": 1},
			wantErr: true,
		},
		{
			name:    "missing timeout",
			config:  &map[string]interface{}{"signal": "ready"},
			wantErr: true,
		},
		{
			name: "empty signal",
			config: &map[string]interface{}{
				"signal":         "",
				"timeoutSeconds": 1,
			},
			wantErr: true,
		},
		{
			name: "invalid signal chars",
			config: &map[string]interface{}{
				"signal":         "approval granted",
				"timeoutSeconds": 1,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &map[string]interface{}{
				"signal":         "ready",
				"timeoutSeconds": -1,
			},
			wantErr: true,
		},
		{
			name: "timeout too large",
			config: &map[string]interface{}{
				"signal":         "ready",
				"timeoutSeconds": maxWaitSeconds + 1,
			},
			wantErr: true,
		},
		{
			name: "unknown property",
			config: &map[string]interface{}{
				"signal":         "ready",
				"timeoutSeconds": 1,
				"extra":          true,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := api.Node{Id: "wait-1", Type: WaitNodeType, Config: test.config}
			err := ValidateNodeConfig(node)
			if test.wantErr && err == nil {
				t.Fatal("ValidateNodeConfig() error = nil, want an error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("ValidateNodeConfig() error = %v", err)
			}
		})
	}
}

func TestWaitTimeoutDuration(t *testing.T) {
	t.Parallel()

	if got := waitTimeoutDuration(api.WaitNodeConfig{TimeoutSeconds: 2.5}).Seconds(); got != 2.5 {
		t.Fatalf("waitTimeoutDuration(2.5) = %v seconds, want 2.5", got)
	}
}
