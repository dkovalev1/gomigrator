package internal

import (
	"database/sql"
	"embed"
	"log"
	"time"

	"github.com/dkovalev1/gomigrator/config"
	_ "github.com/jackc/pgx/v5/stdlib" //nolint
	"github.com/jmoiron/sqlx"          //nolint
	_ "github.com/lib/pq"
)

//go:embed create*.sql
var createTables embed.FS

type Database struct {
	conn *sqlx.DB
}

type MigrationStatus int

const (
	MigrationNew MigrationStatus = iota
	MigrationInProc
	MigrationError
	MigrationApplied
)

func (status MigrationStatus) String() string {
	switch status {
	case MigrationNew:
		return "new"
	case MigrationInProc:
		return "inprogress"
	case MigrationError:
		return "error"
	case MigrationApplied:
		return "applied"
	default:
		panic("Unknown status")
	}
}

type MigrationRec struct {
	Id              int
	Name            string
	MigrationType   config.MigrationType
	MigrationStatus MigrationStatus
	LastUpdated     time.Time
}

func NewDatabase(dsn string) *Database {
	db := &Database{}
	if err := db.init(dsn); err != nil {
		// Nothing to do in the database utility, nothing what we can recover, so just panic here
		panic(err)
	}
	return db
}

func (d *Database) init(dsn string) error {

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}
	d.conn = db

	if isInit, err := d.isMigratorInit(); err != nil {
		return err
	} else if !isInit {
		if err := d.createTables(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) isMigratorInit() (bool, error) {
	result := make([]bool, 0)
	err := d.conn.Select(&result, `
SELECT EXISTS (
   		SELECT FROM information_schema.tables 
   			WHERE  table_schema = 'gomigrator'
   			AND    table_name   = 'migrations')
	`)

	if err != nil {
		return false, err
	}

	return result[0], nil
}

func (d *Database) createTables() error {

	file, err := createTables.Open("create.sql")
	if err != nil {
		return err
	}
	defer file.Close()

	mirgator := NewMigrator(config.Config{}, MigrationUp)

	log.Println("gomigrator initialize. Creating tables...")

	statements, err := mirgator.ReadMigrationStatements(file)
	if err != nil {
		return err
	}

	log.Printf("Executing %d statements\n", len(statements))
	for _, statement := range statements {
		if _, err := d.conn.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) GetVersion() (string, error) {
	var version string
	query := "SELECT version FROM gomigrator.migrations ORDER BY last_run DESC LIMIT 1"
	if err := d.conn.Get(&version, query); err != nil {
		return "", err
	}
	return version, nil
}

func (d *Database) CreateMigration(name string, migrationType config.MigrationType) error {
	sql := `INSERT INTO gomigrator.migrations (name, type) VALUES ($1, $2)`

	_, err := d.conn.Exec(sql, name, migrationType.String())
	return err
}

func (d *Database) GetMigrations() (*[]*MigrationRec, error) {
	sql := `SELECT id, name, migration_type, status, last_run FROM gomigrator.migrations`

	var records []*MigrationRec
	err := d.conn.Select(&records, sql)
	if err != nil {
		return nil, err
	}
	return &records, nil
}

func (d *Database) GetReadyMigrations() (*[]MigrationRec, error) {
	sql := `
SELECT id, name, migration_type, status, last_run 
FROM gomigrator.migrations 
WHERE status = 'new'
ORDER BY name
`

	var records []MigrationRec
	err := d.conn.Select(&records, sql)
	if err != nil {
		return nil, err
	}
	return &records, nil
}

func (d *Database) SetMigrationStatus(mg string, status MigrationStatus) error {
	sql := `UPDATE gomigrator.migrations (name, status) VALUES ($1, $2)`

	_, err := d.conn.Exec(sql, mg, status.String())
	return err
}

func (db *Database) StartTransaction() (tx *sql.Tx) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil
	}
	return
}

func (db *Database) Execute(stmt string) error {
	_, err := db.conn.Exec(stmt)
	return err
}
