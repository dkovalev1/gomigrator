package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/dkovalev1/gomigrator/pkg"
)

func init() {
}

type commandDefinition struct {
	name string
	fn   func(config config.Config) error
	help string
}

var commands = []commandDefinition{
	{"create", pkg.DoCreate, ""},
	{"up", pkg.DoUp, ""},
	{"down", pkg.DoDown, ""},
	{"redo", pkg.DoRedo, ""},
	{"status", pkg.DoStatus, ""},
	{"dbversion", pkg.DoDbversion, ""},
}

var errCommandNotFound = errors.New("command not found")

func usage(err error) {
	fmt.Printf("%v", err)
	fmt.Printf("Commands are:\n")
	for _, cmd := range commands {
		fmt.Printf("    %s  -  %s\n", cmd.name, cmd.help)
	}
	flag.Usage()
}

func runCommand(command string, config config.Config) error {

	fmt.Printf("Performing migration command %s, DSN=%s, migrationPath=%s, migrationType=%s\n",
		command, config.DSN, config.MigrationPath, config.MigrationType.String())

	for _, cmd := range commands {
		if cmd.name == command {
			return cmd.fn(config)
		}
	}
	err := fmt.Errorf("%v %s", errCommandNotFound, command)
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

	runCommand(command, config)
}
