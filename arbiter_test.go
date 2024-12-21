package arbiter

import (
	"errors"
	"os"
	"testing"

	"tres-bon.se/arbiter/pkg/module"
	"tres-bon.se/arbiter/pkg/subcommand/cli"
)

func TestRun_NoSubcommand(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set args to simulate no subcommand given
	os.Args = []string{"arbiter"}

	// Create a dummy module
	modules := module.Modules{&module.MockModule{SetName: "mock"}}

	// Run the function and check for the expected error
	err := Run(modules)
	if err == nil || !errors.Is(err, ErrNoSubcommand) {
		t.Fatalf("expected error: %v, got: %v", ErrNoSubcommand, err)
	}
}

func TestRun_DurationTooShort(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set args to simulate duration flag less than 30 seconds
	os.Args = []string{"arbiter", "-duration", "10s", cli.FlagsetName}

	// Create a dummy module
	modules := module.Modules{&module.MockModule{SetName: "mock"}}

	// Run the function and check for the expected error
	err := Run(modules)
	if err == nil || !errors.Is(err, ErrDurationTooShort) {
		t.Fatalf("expected error: %v, got: %v", ErrDurationTooShort, err)
	}
}
