package gomigrator

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	config "github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoCreate(configuration config.Config, args ...string) error {
	log.Printf("create, migrationType=%s\n", configuration.MigrationType.String())

	if len(args) == 0 {
		return fmt.Errorf("argument <Migration Name> required for create")
	}

	db := internal.NewDatabase(configuration.DSN)
	defer db.Close()

	migrationName := args[0]

	// The migration is either SQL or go.
	// In first case we have a corresponding file in the SQL directory
	// In the case of we expect it to be registered.
	// Let's consider SQL has a priority, as it's easy to override something compiled

	checkPath := path.Join(configuration.MigrationPath, migrationName)
	mType := config.MigrationSQL
	if _, err := os.Stat(checkPath); errors.Is(err, os.ErrNotExist) {
		// SQL file does not exist, check go migration
		if internal.Registry.Check(migrationName) {
			mType = config.MigrationGo
		}
	}

	err := db.CreateMigration(migrationName, mType)

	return err
}
