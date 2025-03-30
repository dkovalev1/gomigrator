package gomigrator

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dkovalev1/gomigrator/config"            //nolint
	internal "github.com/dkovalev1/gomigrator/internal" //nolint
)

type VersionInfo struct {
	Version       int
	MigrationName string
}

func DBVersion(config config.Config) (version VersionInfo, err error) {
	db := internal.NewDatabase(config.DSN)
	defer db.Close()

	dbversion, err := db.GetVersion()
	if err == nil {
		version.Version = dbversion.Version
		version.MigrationName = dbversion.MigrationName
	}

	return
}

func DoDbversion(config config.Config, _ ...string) error {
	fmt.Printf(
		"dbversion, dsn=%s, migrationPath=%s, migrationType=%s\n",
		config.DSN, config.MigrationPath, config.MigrationType.String())

	version, err := DBVersion(config)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}
	if err == nil {
		fmt.Printf("Version: %d migration %s\n", version.Version, version.MigrationName)
	} else {
		fmt.Println("no migrations")
	}

	return nil
}
