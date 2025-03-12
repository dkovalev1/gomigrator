package pkg

import (
	"errors"
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoDbversion(config config.Config) error {
	fmt.Printf("dbversion, dsn=%s, migrationPath=%s, migrationType=%s\n", config.DSN, config.MigrationPath, config.MigrationType.String())

	db := internal.NewDatabase(config.DSN)

	version, err := db.GetVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
	return errors.ErrUnsupported
}
