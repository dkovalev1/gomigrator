package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"   //nolint
	"github.com/dkovalev1/gomigrator/internal" //nolint
)

func DoRedo(config config.Config, _ ...string) error {
	fmt.Printf(
		"redo, dsn=%s, migrationPath=%s, migrationType=%s\n",
		config.DSN, config.MigrationPath, config.MigrationType.String())

	redoMigrator := internal.NewMigrator(config, internal.MigrationDown)
	defer redoMigrator.Close()

	err := redoMigrator.Migrate()
	if err != nil {
		return err
	}

	upMigrator := internal.NewMigrator(config, internal.MigrationUp)
	defer upMigrator.Close()

	err = upMigrator.Migrate()
	if err != nil {
		return err
	}

	return nil
}
