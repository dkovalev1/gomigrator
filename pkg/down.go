package gomigrator

import (
	"github.com/dkovalev1/gomigrator/config"   //nolint
	"github.com/dkovalev1/gomigrator/internal" //nolint
)

func DoDown(config config.Config, _ ...string) error {
	migrator := internal.NewMigrator(config, internal.MigrationDown)
	defer migrator.Close()
	return migrator.Migrate()
}
