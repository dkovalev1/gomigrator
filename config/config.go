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

func NewConfig(configFile string) Config {
	config := Config{
		DSN:           "",
		MigrationPath: "migrations",
		MigrationType: MigrationSQL,
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal(err)
	}

	return config
}
