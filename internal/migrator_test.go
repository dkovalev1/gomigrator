package internal

import (
	"strings"
	"testing"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/stretchr/testify/require"
)

func TestReadMigrationStatements(t *testing.T) {
	migrator := NewMigrator(config.DefaultConfig, MigrationUp)

	t.Run("empty file", func(t *testing.T) {
		ms, err := migrator.ReadMigrationStatements(nil)
		require.Error(t, err)
		require.Nil(t, ms)
	})

	t.Run("invalid file", func(t *testing.T) {
		ms, err := migrator.ReadMigrationStatementsFile("invalid file.txt")
		require.Error(t, err)
		require.Nil(t, ms)
	})

	t.Run("Simple file", func(t *testing.T) {
		test_desc := `
--gomigrator up
CREATE SCHEMA IF NOT EXISTS gomigrator;

--gomigrator up
CREATE TYPE migration_status AS ENUM ('new', 'inprogress', 'error', 'applied');

--gomigrator up
CREATE TYPE migration_type AS ENUM ('go', 'sql');

--gomigrator down
Down
statement
in
few
lines
`
		reader := strings.NewReader(test_desc)

		ms, err := migrator.ReadMigrationStatements(reader)
		require.NoError(t, err)
		require.NotNil(t, ms)
		require.Equal(t, len(ms), 3)

		reader = strings.NewReader(test_desc)
		downMigrator := NewMigrator(config.DefaultConfig, MigrationDown)
		ms, err = downMigrator.ReadMigrationStatements(reader)
		require.NoError(t, err)
		require.NotNil(t, ms)
		require.Equal(t, len(ms), 1)
	})
}
