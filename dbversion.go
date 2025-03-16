package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
	internal "github.com/dkovalev1/gomigrator/internal"
)

func DoDbversion(config config.Config, args []string) error {
	fmt.Printf("dbversion, dsn=%s, migrationPath=%s, migrationType=%s\n", config.DSN, config.MigrationPath, config.MigrationType.String())

	db := internal.NewDatabase(config.DSN)

	version, err := db.GetVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
	return nil
}
