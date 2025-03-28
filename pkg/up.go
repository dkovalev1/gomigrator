package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoUp(cfg config.Config, args []string) error {
	fmt.Printf("up, dsn=%s, migrationPath=%s, migrationType=%s\n", cfg.DSN, cfg.MigrationPath, cfg.MigrationType.String())

	migrator := internal.NewMigrator(cfg, internal.MigrationUp)
	return migrator.Migrate()
}
