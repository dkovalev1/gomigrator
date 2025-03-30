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

func (m *Migrator) Close() {
	if m.Database != nil {
		m.Database.Close()
	}
}

func (m *Migrator) Migrate() error {

	statusFilter := MigrationNew
	if m.Direction == MigrationDown {
		statusFilter = MigrationApplied
	}

	migrations, err := m.Database.GetMigrations(statusFilter)
	if err != nil {
		return err
	}

	log.Printf("Found %d migrations.", len(migrations))
	for _, mg := range migrations {
		err = m.ApplyMigration(&mg)

		if err != nil {
			return err
		}
	}
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

func (m *Migrator) ApplyMigration(mg *MigrationRec) error {
	log.Printf("Apply migration %s to %s\n", mg.Name, m.Direction.String())
	tx := m.Database.StartTransaction()

	var status_err error = nil
	var err error = nil
	defer func() {
		// Migration can be failed bus status shall be updated.
		// We do it in the same transaction, commit shall be succeeded.
		tx.Rollback()

		if status_err != nil || err != nil {
			// Update status in the different transaction
			status_err = m.Database.SetMigrationStatus(mg.Id, MigrationError)
		}
	}()

	status := MigrationError

	switch mg.Type {
	case config.MigrationSQL:
		err = m.applySqlMigration(tx, mg)
	case config.MigrationGo:
		migration := Registry.Get(mg.Name)
		err = m.applyGoMigration(tx, migration)
	}

	if err != nil {
		log.Printf("Error %s in migration %s: %v", m.Direction.String(), mg.Name, err)
	} else {
		log.Printf("migration %s applied", mg.Name)

		if m.Direction == MigrationDown {
			status = MigrationNew
		} else {
			status = MigrationApplied
		}
	}

	if err == nil {
		status_err = m.Database.SetMigrationStatus(mg.Id, status)
		err = tx.Commit()
	} else {
		log.Printf("Error %s in migration %s: %v", m.Direction.String(), mg.Name, err)
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

	checkPath := path.Join(m.Config.MigrationPath, mg.Name+".sql")

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
