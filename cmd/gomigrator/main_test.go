package main

import (
	"errors"
	"testing"

	"github.com/dkovalev1/gomigrator/config" //nolint
	"github.com/stretchr/testify/require"    //nolint
)

var (
	errTestError = errors.New("test")
	testConfig   = config.Config{
		DSN: "test dsn",
	}
)

func TestRunCommand(t *testing.T) {
	wasCalled := false

	runTest := func(config config.Config, args ...string) error {
		wasCalled = true
		require.Equal(t, "test dsn", config.DSN)
		require.Equal(t, len(args), 2)
		require.Equal(t, "arg0", args[0])
		require.Equal(t, "arg1", args[1])
		// okay
		return errTestError
	}

	commands["test"] = commandDefinition{
		fn:   runTest,
		help: "just for test",
	}

	err := runCommand("test", testConfig, []string{"arg0", "arg1"})
	require.Equal(t, err, errTestError)

	require.True(t, wasCalled)
}
