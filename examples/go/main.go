package main

import (
	"flag"

	"github.com/dkovalev1/gomigrator/config"
	gomigrator "github.com/dkovalev1/gomigrator/pkg"
)

func Migrate() {
	config := config.Config{}
	if err := gomigrator.DoUp(config); err != nil {
		panic(err)
	}
}

func main() {

	flag.Parse()

	Migrate()
}
