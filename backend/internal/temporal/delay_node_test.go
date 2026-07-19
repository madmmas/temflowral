package temporal

import (
	"testing"

	"github.com/madmmas/temflowral/backend/internal/api"
)

func TestValidateDelayNodeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *map[string]interface{}
		wantErr bool
	}{
		{name: "valid", config: &map[string]interface{}{"seconds": 5}},
		{name: "zero", config: &map[string]interface{}{"seconds": 0}},
		{name: "fractional", config: &map[string]interface{}{"seconds": 0.25}},
		{name: "missing config", config: nil, wantErr: true},
		{name: "missing seconds", config: &map[string]interface{}{}, wantErr: true},
		{name: "negative", config: &map[string]interface{}{"seconds": -1}, wantErr: true},
		{name: "too large", config: &map[string]interface{}{"seconds": maxDelaySeconds + 1}, wantErr: true},
		{name: "unknown property", config: &map[string]interface{}{"seconds": 1, "until": "later"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := api.Node{Id: "delay-1", Type: DelayNodeType, Config: test.config}
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

func TestDelayDuration(t *testing.T) {
	t.Parallel()

	if got := delayDuration(api.DelayNodeConfig{Seconds: 2.5}).Seconds(); got != 2.5 {
		t.Fatalf("delayDuration(2.5) = %v seconds, want 2.5", got)
	}
}
