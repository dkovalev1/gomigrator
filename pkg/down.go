package pkg

import (
	"errors"
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
)

func DoDown(config config.Config) error {
	fmt.Printf("down, dsn=%s, migrationPath=%s, migrationType=%s\n", config.DSN, config.MigrationPath, config.MigrationType.String())
	return errors.ErrUnsupported
}
