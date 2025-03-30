package gomigrator

import (
	"fmt"
	"time"

	"github.com/dkovalev1/gomigrator/config"   //nolint
	"github.com/dkovalev1/gomigrator/internal" //nolint
)

type MigrationStatusRec struct {
	ID      int
	Name    string
	Type    string
	Status  string
	LastRun time.Time
	Applied bool
}

func Status(config config.Config) (status []MigrationStatusRec, err error) {
	database := internal.NewDatabase(config.DSN)
	defer database.Close()

	migrations, err := database.GetMigrations()
	if err != nil {
		return
	}

	for _, m := range migrations {
		rec := MigrationStatusRec{
			ID:      m.ID,
			Name:    m.Name,
			Type:    m.Type.String(),
			Status:  m.Status.String(),
			LastRun: m.LastRun,
			Applied: m.Applied,
		}
		status = append(status, rec)
	}

	return
}

func DoStatus(config config.Config, _ ...string) error {
	fmt.Printf(
		"status, dsn=%s, migrationPath=%s, migrationType=%s\n",
		config.DSN, config.MigrationPath, config.MigrationType.String())

	migrations, err := Status(config)
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
		fmt.Printf("%6d|%10s|%6s|%10s|%15s\n", m.ID, m.Name, m.Type, m.Status, applied)
	}
	return nil
}
