package gomigrator

import (
	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/internal"
)

func DoDown(config config.Config, args ...string) error {
	migrator := internal.NewMigrator(config, internal.MigrationDown)
	defer migrator.Close()
	return migrator.Migrate()
}
