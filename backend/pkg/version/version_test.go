package version_test

import (
	"testing"

	"github.com/madmmas/temflowral/backend/pkg/version"
)

func TestModuleAndVersion(t *testing.T) {
	if version.Module == "" {
		t.Fatal("Module is empty")
	}
	if version.Version == "" {
		t.Fatal("Version is empty")
	}
}
