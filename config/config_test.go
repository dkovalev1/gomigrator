package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDefault(t *testing.T) {
	cfg := NewConfig("invalid.txt")
	require.NotNil(t, cfg)
	require.Equal(t, DefaultConfig.DSN, cfg.DSN)
	require.Equal(t, DefaultConfig.MigrationPath, cfg.MigrationPath)
	require.Equal(t, DefaultConfig.MigrationType, cfg.MigrationType)
}
