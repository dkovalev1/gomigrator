package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dkovalev1/gomigrator/config"         //nolint
	gomigrator "github.com/dkovalev1/gomigrator/pkg" //nolint
)

type commandDefinition struct {
	fn   func(config config.Config, args ...string) error
	help string
}

var commands = map[string]commandDefinition{
	"create":    {gomigrator.DoCreate, "create a migration"},
	"up":        {gomigrator.DoUp, "apply all migrations"},
	"down":      {gomigrator.DoDown, "revert migrations"},
	"redo":      {gomigrator.DoRedo, "redo latest migration (undo + redo)"},
	"status":    {gomigrator.DoStatus, "show migration status"},
	"dbversion": {gomigrator.DoDbversion, "show latest migration"},
}

var errCommandNotFound = errors.New("command not found")

func usage(err error) {
	fmt.Printf("%v", err)
	fmt.Printf("Commands are:\n")
	for name, cmd := range commands {
		fmt.Printf("    %s  -  %s\n", name, cmd.help)
	}
	flag.Usage()
}

func runCommand(command string, config config.Config, args []string) error {
	fmt.Printf("Performing migration command %s, migrationPath=%s, migrationType=%s\n",
		command, config.MigrationPath, config.MigrationType.String())

	cmd, ok := commands[command]
	if ok {
		return cmd.fn(config, args...)
	}

	err := fmt.Errorf("%w %s", errCommandNotFound, command)
	usage(err)
	return err
}

func main() {
	var configFile string

	configCommand := flag.NewFlagSet("config", flag.ExitOnError)
	configCommand.StringVar(&configFile, "config", "config.toml", "config file")
	configCommand.Parse(os.Args)

	config := config.NewConfig(configFile)

	flag.StringVar(&config.DSN, "DSN", config.DSN, "Data Source Name")
	flag.StringVar(&config.MigrationPath, "path", config.MigrationPath, "path to the migrations")
	flag.Var(&config.MigrationType, "type", "migration type")

	flag.Parse()

	if !flag.Parsed() {
		fmt.Println("Failed to parse flags")
		os.Exit(1)
	}

	if flag.NArg() < 1 {
		fmt.Println("Missing command")
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	err := runCommand(command, config, flag.Args()[1:])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Ok.")
	os.Exit(0)
}
