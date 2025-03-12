package internal

import (
	//nolint
	"bufio"
	"embed"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" //nolint
	"github.com/jmoiron/sqlx"          //nolint
)

//go:embed create*.sql
var createTables embed.FS

type Database struct {
	conn *sqlx.DB
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
   			AND    table_name   = '')
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

	log.Println("gomigrator initialize. Creating tables...")

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	statement := ""
	statements := make([]string, 0)
	for scanner.Scan() {

		line := scanner.Text()
		if strings.HasPrefix(line, "--gomigrator") {
			statements = append(statements, statement)
			statement = ""
		} else {
			statement += line
		}

		if _, err := d.conn.Exec(scanner.Text()); err != nil {
			return err
		}
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
	if err := d.conn.Get(&version, "SELECT version FROM gomigrator.migrations ORDER BY last_run DESC LIMIT 1"); err != nil {
		return "", err
	}
	return version, nil
}
