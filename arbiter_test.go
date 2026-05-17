package arbiter

import (
	"os"
	"testing"

	"github.com/maansaake/arbiter/pkg/module"
	modulemock "github.com/maansaake/arbiter/pkg/module/mock"
	"github.com/maansaake/arbiter/pkg/subcommand/cli"
)

func TestRun_DurationTooShort(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set args to simulate duration flag less than 1 second
	os.Args = []string{"arbiter", "--duration", "0s", cli.FlagsetName}

	// Create a dummy module
	modules := module.Modules{&modulemock.Module{SetName: "mock"}}

	// Run the function and check for the expected error
	err := Run(modules, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRun_ParsingFailed(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	t.Run("empty report path", func(t *testing.T) {
		// Set args to simulate duration flag less than 1 second
		os.Args = []string{"arbiter", "-d", "1s", "--report-path", "", cli.FlagsetName}

		modules := module.Modules{&modulemock.Module{SetName: "mock"}}

		err := Run(modules, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("report path is a directory", func(t *testing.T) {
		// Create a temporary directory to use as the report path
		dir := t.TempDir()

		os.Args = []string{"arbiter", "-d", "1s", "--report-path", dir, cli.FlagsetName}

		modules := module.Modules{&modulemock.Module{SetName: "mock"}}

		err := Run(modules, nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
