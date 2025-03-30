package main

import (
	"errors"
	"testing"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/stretchr/testify/require"
)

var errTestError = errors.New("test")

func TestRunCommand(t *testing.T) {
	was_called := false
	test_config := config.Config{
		DSN: "test dsn",
	}
	runTest := func(config config.Config, args ...string) error {
		was_called = true
		require.Equal(t, "test dsn", config.DSN)
		require.Equal(t, len(args), 2)
		require.Equal(t, "arg0", args[0])
		require.Equal(t, "arg1", args[1])
		// okay
		return errTestError
	}
	test_cmd := commandDefinition{
		name: "test",
		fn:   runTest,
		help: "just for test",
	}
	commands = append(commands, test_cmd)

	err := runCommand("test", test_config, []string{"arg0", "arg1"})
	require.Equal(t, err, errTestError)

	require.True(t, was_called)
}
