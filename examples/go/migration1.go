package main

import (
	"database/sql"
	"errors"

	gomigrator "github.com/dkovalev1/gomigrator/pkg"
)

func init() {
	gomigrator.Register("mig2go", up, down)
}

func up(tx *sql.Tx) error {
	return errors.ErrUnsupported
}

func down(tc *sql.Tx) error {
	return errors.ErrUnsupported
}
