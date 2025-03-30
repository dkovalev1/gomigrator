package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoRedo(config config.Config, args ...string) error {
	fmt.Printf("redo, dsn=%s, migrationPath=%s, migrationType=%s\n", config.DSN, config.MigrationPath, config.MigrationType.String())

	redoMigrator := internal.NewMigrator(config, internal.MigrationDown)
	defer redoMigrator.Close()

	err := redoMigrator.Migrate()
	if err != nil {
		return err
	}

	upMigrator := internal.NewMigrator(config, internal.MigrationUp)
	err = upMigrator.Migrate()
	if err != nil {
		return err
	}

	return nil
}
