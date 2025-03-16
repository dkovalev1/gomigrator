package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoStatus(config config.Config, args []string) error {
	fmt.Printf("status, dsn=%s, migrationPath=%s, migrationType=%s\n", config.DSN, config.MigrationPath, config.MigrationType.String())

	database := internal.NewDatabase(config.DSN)
	migrations, err := database.GetMigrations()
	if err != nil {
		return err
	}

	for _, m := range *migrations {
		fmt.Printf("%s\n", m.Name)
	}
	return nil
}
