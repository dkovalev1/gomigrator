package main

import (
	"database/sql"
	"errors"

	gom "github.com/dkovalev1/gomigrator"
)

func init() {
	gom.Register("mig2go", up, down)
}

func up(tx *sql.Tx) error {
	return errors.ErrUnsupported
}

func down(tc *sql.Tx) error {
	return errors.ErrUnsupported
}
