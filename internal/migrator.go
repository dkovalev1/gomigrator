package internal

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/dkovalev1/gomigrator/config"
)

type MigrationDirection int

const (
	MigrationUp = iota
	MigrationDown
)

type Migrator struct {
	Config    config.Config
	Direction MigrationDirection
	Database  *Database
}

func (md MigrationDirection) String() string {
	switch md {
	case MigrationUp:
		return "UP"
	case MigrationDown:
		return "DOWN"
	default:
		panic(fmt.Sprintf("Unknown migration direction %d", md))
	}
}

func NewMigrator(cfg config.Config, dir MigrationDirection) *Migrator {
	ret := &Migrator{
		Config:    cfg,
		Direction: dir,
	}
	db := NewDatabase(cfg.DSN)
	ret.Database = db
	return ret
}

func (m *Migrator) Migrate() error {
	tx := m.Database.StartTransaction()
	success := false
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	migrations, err := m.Database.GetMigrations()
	if err != nil {
		return err
	}

	for _, mg := range *migrations {
		err = m.ApplyMigration(tx, mg)

		if err != nil {
			return err
		}
	}
	success = true
	return nil
}

func (m *Migrator) ReadMigrationStatementsFile(filename string) ([]string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return m.ReadMigrationStatements(fp)
}

func (m *Migrator) ReadMigrationStatements(reader io.Reader) ([]string, error) {
	if reader == nil {
		return nil, fmt.Errorf("invalid reader")
	}
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	statement := ""
	statements := make([]string, 0)
	stmtStarted := false

	var currentDirection MigrationDirection = MigrationUp

	for scanner.Scan() {

		line := scanner.Text()
		if strings.HasPrefix(line, "--gomigrator up") {
			if stmtStarted && currentDirection == m.Direction {
				statements = append(statements, statement)
			}
			statement = ""
			stmtStarted = true
			currentDirection = MigrationUp
		} else if strings.HasPrefix(line, "--gomigrator down") {
			if stmtStarted && currentDirection == m.Direction {
				statements = append(statements, statement)
			}
			statement = ""
			stmtStarted = true
			currentDirection = MigrationDown
		} else if line != "" {
			statement += line
		}
	}
	if stmtStarted && statement != "" && currentDirection == m.Direction {
		statements = append(statements, statement)
	}

	return statements, nil
}

func (m *Migrator) ApplyMigration(tx *sql.Tx, mg *MigrationRec) error {
	migration := Registry.Get(mg.Name)

	var err error
	switch mg.MigrationType {
	case config.MigrationSQL:
		m.applySqlMigration(tx, mg)
	case config.MigrationGo:
		err = m.applyGoMigration(tx, migration)
	}

	if err != nil {
		log.Printf("Error in %s for migration %s: %v", m.Direction.String(), mg.Name, err)
	} else {
		err = tx.Commit()
		if err != nil {
			log.Printf("Error in COMMIT for %s migration %s: %v", m.Direction.String(), mg.Name, err)
		} else {
			log.Printf("migration %s applied", mg.Name)
		}
	}
	status := MigrationApplied
	if m.Direction == MigrationDown {
		status = MigrationNew
	}
	if err != nil {
		status = MigrationError
	}

	err = m.Database.SetMigrationStatus(mg.Name, status)
	if err != nil {
		panic(fmt.Sprintf("Migration %s applied, but I can not update status: %v", mg.Name, err))
	}
	return err
}

func (m *Migrator) applyGoMigration(tx *sql.Tx, migration *Migration) error {
	var err error

	switch m.Direction {
	case MigrationUp:
		err = migration.Up(tx)
	case MigrationDown:
		err = migration.Down(tx)
	default:
		panic(fmt.Sprintf("Unknown migration direction %d", m.Direction))
	}

	return err
}

func (m *Migrator) applySqlMigration(tx *sql.Tx, mg *MigrationRec) error {

	checkPath := path.Join(m.Config.MigrationPath, mg.Name)

	_, err := os.Stat(checkPath)
	if err != nil {
		return fmt.Errorf("file not found %s", checkPath)
	}

	file, err := os.Open(checkPath)
	if err != nil {
		return err
	}
	defer file.Close()

	statements, err := m.ReadMigrationStatements(file)

	if err != nil {
		return nil
	}

	for _, stmt := range statements {
		_, err := tx.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}
