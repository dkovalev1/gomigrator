package internal

import (
	"database/sql"
)

type UpMigration func(sql.Tx) error
type DownMigration func(sql.Tx) error

type IRegistry interface {
	Register(name string, up func(Tx *sql.Tx) error, down func(Tx *sql.Tx) error)
	Check(name string) bool
	Get(name string) *Migration
}

var Registry IRegistry

type Migration struct {
	Name string
	Up   func(Tx *sql.Tx) error
	Down func(Tx *sql.Tx) error
}

type RegistryImpl struct {
	records map[string]Migration
}

func init() {
	Registry = NewRegistry()
}

func NewRegistry() IRegistry {
	return &RegistryImpl{
		records: make(map[string]Migration),
	}
}

func (r *RegistryImpl) Register(name string, up func(Tx *sql.Tx) error, down func(Tx *sql.Tx) error) {
	m := Migration{
		Name: name,
		Up:   up,
		Down: down,
	}

	r.records[name] = m
}

func (r *RegistryImpl) Check(name string) bool {
	_, ok := r.records[name]
	return ok
}

func (r *RegistryImpl) Get(name string) *Migration {
	val, ok := r.records[name]
	if ok {
		return &val
	}
	return nil
}
