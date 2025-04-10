package internal

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/dkovalev1/gomigrator/config" //nolint
	_ "github.com/jackc/pgx/v5/stdlib"       //nolint
	"github.com/jmoiron/sqlx"                //nolint
	_ "github.com/lib/pq"                    //nolint
)

//go:embed create*.sql
var createTables embed.FS

type Database struct {
	conn *sqlx.DB
}

type OrderBy int

const (
	OrderByNone OrderBy = iota
	OrderByAsc
	OrderByDesc
)

type MigrationStatus int

const (
	MigrationNew MigrationStatus = iota
	MigrationInProc
	MigrationError
	MigrationApplied
)

var errInvalidMigrationStatus = fmt.Errorf("invalid migration status")

func (s *MigrationStatus) Set(value string) error {
	switch value {
	case "new":
		*s = MigrationNew
	case "inprogress":
		*s = MigrationInProc
	case "error":
		*s = MigrationError
	case "applied":
		*s = MigrationApplied
	default:
		return errInvalidMigrationStatus
	}
	return nil
}

func (s *MigrationStatus) String() string {
	switch *s {
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
	ID      int
	Name    string
	Type    config.MigrationType
	Status  MigrationStatus
	LastRun time.Time
	Applied bool
}

func NewDatabase(dsn string) *Database {
	db := &Database{}
	if err := db.init(dsn); err != nil {
		// Nothing to do in the database utility, nothing what we can recover, so just panic here
		log.Fatal(err)
	}
	// Acquire lock. The lock lasts till end of the session, no additional actions required
	qlock := "SELECT pg_advisory_lock(('x' || md5('gomigrator_session'))::bit(64)::bigint)"
	_, err := db.conn.Exec(qlock)
	if err != nil {
		log.Fatal(err)
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

/*
Create internal tables for migrator. Use temporary migrator instance for the tables.
*/
func (d *Database) createTables() error {
	file, err := createTables.Open("create.sql")
	if err != nil {
		return err
	}
	defer file.Close()

	mirgator := &Migrator{
		Direction: MigrationUp,
		Database:  d,
	}

	log.Println("gomigrator initialize. Creating tables...")

	statements, err := mirgator.ReadMigrationStatements(file)
	if err != nil {
		return err
	}

	log.Printf("Executing %d statements\n", len(statements))
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range statements {
		if _, err := d.conn.Exec(statement); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type VersionInfo struct {
	Version       int
	MigrationName string
}

func (d *Database) GetVersion() (version VersionInfo, err error) {
	query := `
SELECT mid AS Version, mname AS MigrationName 
FROM gomigrator.migrations 
WHERE mstatus='applied' 
ORDER BY mlastrun 
DESC LIMIT 1`

	if err = d.conn.Get(&version, query); err != nil {
		return VersionInfo{}, err
	}
	return version, nil
}

func (d *Database) CreateMigration(name string, migrationType config.MigrationType) error {
	sql := `INSERT INTO gomigrator.migrations (mname, mtype) VALUES ($1, $2)`

	_, err := d.conn.Exec(sql, name, migrationType.String())
	return err
}

type dbrec struct {
	mid      int
	mname    string
	mtype    string
	mstatus  string
	mlastrun sql.Null[time.Time]
}

func (d *Database) GetMigrations(args ...any) ([]MigrationRec, error) {
	sql := `SELECT mid, mname, mtype, mstatus, mlastrun FROM gomigrator.migrations`
	var ret []MigrationRec
	parameters := make([]any, 0)

	var order string

	for _, arg := range args {
		switch v := arg.(type) {
		case MigrationStatus:
			sql += " WHERE mstatus = $1"
			parameters = append(parameters, v.String())
		case OrderBy:
			if v == OrderByDesc {
				order = " DESC"
			}
		default:
			return nil, fmt.Errorf("invalid argument to GetMigrations")
		}
	}

	sql += " ORDER BY mid" + order

	rows, err := d.conn.Query(sql, parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rec dbrec
		err = rows.Scan(&rec.mid, &rec.mname, &rec.mtype, &rec.mstatus, &rec.mlastrun)
		if err != nil {
			return nil, err
		}
		mrec := MigrationRec{
			ID:   rec.mid,
			Name: rec.mname,
		}
		if rec.mlastrun.Valid {
			mrec.LastRun = rec.mlastrun.V
			mrec.Applied = true
		}
		mrec.Type.Set(rec.mtype)
		mrec.Status.Set(rec.mstatus)
		ret = append(ret, mrec)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (d *Database) GetReadyMigrations() (*[]MigrationRec, error) {
	sql := `
SELECT mid, mname, mtype, mstatus, mlastrun
FROM gomigrator.migrations 
WHERE status = 'new'
ORDER BY name
`

	var records []dbrec
	err := d.conn.Select(&records, sql)
	if err != nil {
		return nil, err
	}
	ret := make([]MigrationRec, 0, len(records))
	for _, rec := range records {
		mrec := MigrationRec{
			ID:   rec.mid,
			Name: rec.mname,
		}
		if rec.mlastrun.Valid {
			mrec.LastRun = rec.mlastrun.V
			mrec.Applied = true
		}
		mrec.Type.Set(rec.mtype)
		mrec.Status.Set(rec.mstatus)
		ret = append(ret, mrec)
	}
	return &ret, nil
}

func (d *Database) SetMigrationStatus(mid int, status MigrationStatus) error {
	sql := `UPDATE gomigrator.migrations SET mstatus=$1, mlastrun=NOW() WHERE mid=$2`
	_, err := d.conn.Exec(sql, status.String(), mid)
	return err
}

func (d *Database) StartTransaction() (tx *sql.Tx) {
	tx, err := d.conn.Begin()
	if err != nil {
		return nil
	}
	return
}

func (d *Database) Execute(stmt string) error {
	_, err := d.conn.Exec(stmt)
	return err
}
