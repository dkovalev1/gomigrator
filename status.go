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

	fmt.Printf("%6s|%10s|%6s|%10s|%15s\n", "id", "name", "type", "status", "last run")
	for range 51 {
		fmt.Print("-")
	}
	fmt.Println("")
	for _, m := range migrations {
		var applied string
		if m.Applied {
			applied = m.LastRun.String()
		} else {
			applied = "NEVER"
		}
		fmt.Printf("%6d|%10s|%6s|%10s|%15s\n", m.Id, m.Name, m.Type.String(), m.Status.String(), applied)
	}
	return nil
}
