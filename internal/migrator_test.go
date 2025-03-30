package internal

import (
	"strings"
	"testing"

	"github.com/dkovalev1/gomigrator/config" //nolint
	"github.com/stretchr/testify/require"    //nolint
)

func TestReadMigrationStatements(t *testing.T) {
	migrator := NewMigrator(config.Config{DSN: "skip"}, MigrationUp)
	defer migrator.Close()

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
		testDesc := `
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
		reader := strings.NewReader(testDesc)

		ms, err := migrator.ReadMigrationStatements(reader)
		require.NoError(t, err)
		require.NotNil(t, ms)
		require.Equal(t, len(ms), 3)

		reader = strings.NewReader(testDesc)

		downMigrator := NewMigrator(config.Config{DSN: "skip"}, MigrationDown)
		defer downMigrator.Close()

		ms, err = downMigrator.ReadMigrationStatements(reader)
		require.NoError(t, err)
		require.NotNil(t, ms)
		require.Equal(t, len(ms), 1)
	})
}
