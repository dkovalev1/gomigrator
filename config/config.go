package config

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml" //nolint
)

type MigrationType int

const (
	MigrationSQL MigrationType = iota
	MigrationGo
)

var errInvalidMigrationType = fmt.Errorf("invalid migration type")

// Set implements flag.Value.
func (m *MigrationType) Set(value string) error {
	switch value {
	case "sql":
		*m = MigrationSQL
	case "go":
		*m = MigrationGo
	default:
		return errInvalidMigrationType
	}
	return nil
}

// String implements flag.Value.
func (m *MigrationType) String() string {
	switch *m {
	case MigrationSQL:
		return "sql"
	case MigrationGo:
		return "go"
	}
	panic("invalid migration type")
}

type Config struct {
	DSN           string
	MigrationPath string
	MigrationType MigrationType
}

var DefaultConfig = Config{
	DSN:           "host=localhost user=test password=test dbname=migratordb sslmode=disable",
	MigrationPath: "migrations",
	MigrationType: MigrationSQL,
}

func NewConfig(configFile string) Config {

	config := DefaultConfig

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Printf("%v. Using default values", err)
	}

	return config
}
