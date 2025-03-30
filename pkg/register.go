package gomigrator

import (
	"database/sql"

	"github.com/dkovalev1/gomigrator/internal"
)

func Register(name string, up func(Tx *sql.Tx) error, down func(Tx *sql.Tx) error) error {
	return internal.Registry.Register(name, up, down)
}
