package gomigrator

import (
	"fmt"

	"github.com/dkovalev1/gomigrator/config"   //nolint
	"github.com/dkovalev1/gomigrator/internal" //nolint
)

func DoUp(cfg config.Config, _ ...string) error {
	fmt.Printf("up, dsn=%s, migrationPath=%s, migrationType=%s\n", cfg.DSN, cfg.MigrationPath, cfg.MigrationType.String())

	migrator := internal.NewMigrator(cfg, internal.MigrationUp)
	defer migrator.Close()

	return migrator.Migrate()
}
