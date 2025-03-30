package integration_test

import (
	"database/sql"
	"fmt"
	"os/exec"
	"testing"

	"github.com/dkovalev1/gomigrator/config"         //nolint
	gomigrator "github.com/dkovalev1/gomigrator/pkg" //nolint
	"github.com/jmoiron/sqlx"                        //nolint
	. "github.com/onsi/ginkgo/v2"                    //nolint
	. "github.com/onsi/gomega"                       //nolint
)

const (
	testDSN        = "host=localhost user=test password=test dbname=migratordb sslmode=disable"
	testExec       = "../gomigrator"
	one      int64 = 1
	two      int64 = 2
)

var testConfig = config.Config{
	DSN:           "host=localhost user=test password=test dbname=migratordb sslmode=disable",
	MigrationType: config.MigrationGo,
	MigrationPath: "migrations",
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func cleanupDatabase() {
	conn, err := sqlx.Connect("postgres", testDSN)
	Expect(err).NotTo(HaveOccurred())

	defer conn.Close()

	// Ensure the connection is actually working
	err = conn.Ping()
	Expect(err).NotTo(HaveOccurred())

	// Delete test assets

	_, err = conn.Exec("DROP TABLE IF EXISTS test1")
	Expect(err).NotTo(HaveOccurred())

	_, err = conn.Exec("DROP TABLE IF EXISTS test2")
	Expect(err).NotTo(HaveOccurred())

	_, err = conn.Exec("DROP TABLE IF EXISTS apitest1")
	Expect(err).NotTo(HaveOccurred())

	_, err = conn.Exec("DROP TABLE IF EXISTS apitest2")
	Expect(err).NotTo(HaveOccurred())

	// Delete migrator assets

	_, err = conn.Exec("DROP TABLE IF EXISTS gomigrator.migrations")
	Expect(err).NotTo(HaveOccurred())

	_, err = conn.Exec("DROP TYPE IF EXISTS migration_type")
	Expect(err).NotTo(HaveOccurred())

	_, err = conn.Exec("DROP TYPE IF EXISTS migration_status")
	Expect(err).NotTo(HaveOccurred())
}

// records from the database for test purposes.
type MigrationRec struct {
	Mid      int
	Mname    string
	Mtype    string
	Mstatus  string
	Mlastrun sql.NullTime
}

func setUpMigrations() {
	err := gomigrator.DoCreate(testConfig, "api1")
	Expect(err).NotTo(HaveOccurred())

	err = gomigrator.DoCreate(testConfig, "api2")
	Expect(err).NotTo(HaveOccurred())

	err = gomigrator.DoCreate(testConfig, "mig1")
	Expect(err).NotTo(HaveOccurred())

	err = gomigrator.DoUp(testConfig)
	Expect(err).NotTo(HaveOccurred())

	tables := getTables()
	Expect(tables).To(ContainElement("apitest1"))
	Expect(tables).To(ContainElement("apitest2"))
	Expect(tables).To(ContainElement("test1"))
}

func checkMigrationsStatus(status []string) {
	hasMigrations := getMigrationRecords()
	Expect(len(hasMigrations)).Should(Equal(3))
	Expect(hasMigrations[0].Mname).Should(Equal("api1"))
	Expect(hasMigrations[0].Mstatus).Should(Equal(status[0]))
	Expect(hasMigrations[1].Mname).Should(Equal("api2"))
	Expect(hasMigrations[1].Mstatus).Should(Equal(status[1]))
	Expect(hasMigrations[2].Mname).Should(Equal("mig1"))
	Expect(hasMigrations[2].Mstatus).Should(Equal(status[2]))
}

func getMigrationRecords() []MigrationRec {
	query := `SELECT mid, mname, mtype, mstatus, mlastrun FROM gomigrator.migrations ORDER BY mid`

	conn, err := sqlx.Connect("postgres", testDSN)
	Expect(err).NotTo(HaveOccurred())

	defer conn.Close()
	// Ensure the connection is actually working
	err = conn.Ping()
	Expect(err).NotTo(HaveOccurred())

	ret := []MigrationRec{}

	err = conn.Select(&ret, query)
	Expect(err).NotTo(HaveOccurred())

	return ret
}

func selectValue(table, field string) (ret any) {
	query := fmt.Sprintf("SELECT %s FROM %s LIMIT 1", field, table)

	conn, err := sqlx.Connect("postgres", testDSN)
	Expect(err).NotTo(HaveOccurred())

	defer conn.Close()
	// Ensure the connection is actually working
	err = conn.Ping()
	Expect(err).NotTo(HaveOccurred())

	err = conn.Get(&ret, query)
	Expect(err).NotTo(HaveOccurred())

	return ret
}

func getTables() (tables []string) {
	query := "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'"

	conn, err := sqlx.Connect("postgres", testDSN)
	Expect(err).NotTo(HaveOccurred())
	defer conn.Close()

	err = conn.Select(&tables, query)
	Expect(err).NotTo(HaveOccurred())

	return
}

func checkOutput(args ...string) string {
	out, err := exec.Command(testExec, args...).CombinedOutput()
	strout := string(out)

	Expect(err).NotTo(HaveOccurred(), strout)

	Expect(strout).Should(ContainSubstring("Ok."))

	return strout
}

var _ = Describe("Integration tests for utility", func() {
	BeforeEach(func() {
		cleanupDatabase()
	})

	AfterEach(func() {
		// cleanupDatabase()
	})

	It("creates migration", func() {
		checkOutput("create", "mig1")

		// Check that mig1 is in migrator table
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(1))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
	})

	It("gets status", func() {
		checkOutput("create", "mig1")

		out := checkOutput("status")
		Expect(out).Should(ContainSubstring("mig1"))
	})

	It("dbversion", func() {
		checkOutput("create", "mig1")

		// Check that migrations not applied
		out := checkOutput("dbversion")
		Expect(out).Should(ContainSubstring("no migrations"))

		checkOutput("up")
		out = checkOutput("dbversion")
		Expect(out).Should(ContainSubstring("1"))
	})

	It("applies single migration - up", func() {
		checkOutput("create", "mig1")
		checkOutput("up")

		// Check that mig1 is in migrator table and status is applied
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(1))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))

		value := selectValue("test1", "i")
		Expect(value).Should(Equal(one))
	})

	It("applies migrations - up", func() {
		checkOutput("create", "mig1")
		checkOutput("create", "mig2")
		checkOutput("up")

		// Check that mig1 is in migrator table and status is applied
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(2))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))

		value1 := selectValue("test1", "i")
		Expect(value1).Should(Equal(one))

		Expect(hasMigrations[1].Mname).Should(Equal("mig2"))
		Expect(hasMigrations[1].Mstatus).Should(Equal("applied"))

		value2 := selectValue("test2", "key")
		Expect(value2).Should(Equal("one"))
	})

	It("reverts migrations - down", func() {
		checkOutput("create", "mig1")
		checkOutput("create", "mig2")
		checkOutput("up")

		// make sure migrations applied
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(2))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))
		Expect(hasMigrations[1].Mname).Should(Equal("mig2"))
		Expect(hasMigrations[1].Mstatus).Should(Equal("applied"))

		out := checkOutput("down")

		Expect(out).Should(ContainSubstring("Found 2 migrations."))
		Expect(out).Should(ContainSubstring("Apply migration mig1 to DOWN"))
		Expect(out).Should(ContainSubstring("Apply migration mig2 to DOWN"))

		// Check that mig1 and mig2 are in migrator table and status is new
		hasMigrations = getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(2))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("new"))
		Expect(hasMigrations[1].Mname).Should(Equal("mig2"))
		Expect(hasMigrations[1].Mstatus).Should(Equal("new"))
	})

	It("reapplies migrations - redo", func() {
		checkOutput("create", "mig1")
		checkOutput("create", "mig2")
		checkOutput("up")

		out := checkOutput("redo")
		Expect(out).Should(ContainSubstring("Found 2 migrations."))

		Expect(out).Should(ContainSubstring("Apply migration mig1 to DOWN"))
		Expect(out).Should(ContainSubstring("Apply migration mig2 to DOWN"))

		Expect(out).Should(ContainSubstring("Apply migration mig1 to UP"))
		Expect(out).Should(ContainSubstring("Apply migration mig2 to UP"))

		// Check that mig1 is in migrator table and status is applied
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(2))
		Expect(hasMigrations[0].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))
		Expect(hasMigrations[1].Mname).Should(Equal("mig2"))
		Expect(hasMigrations[1].Mstatus).Should(Equal("applied"))
	})
})

func api1Up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE apitest1(i INT PRIMARY KEY, j INT)")
	if err == nil {
		_, err = tx.Exec("INSERT INTO apitest1(i) VALUES(2)")
	}
	return err
}

func api1Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS apitest1")
	return err
}

func api2Up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE apitest2(\"key\" VARCHAR PRIMARY KEY, j INT)")
	if err == nil {
		_, err = tx.Exec("INSERT INTO apitest2(\"key\", j) VALUES('two', 1)")
	}
	return err
}

func api2Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS apitest2")
	return err
}

var _ = Describe("Integration API tests", func() {
	err := gomigrator.Register("api1", api1Up, api1Down)
	Expect(err).NotTo(HaveOccurred())

	err = gomigrator.Register("api2", api2Up, api2Down)
	Expect(err).NotTo(HaveOccurred())

	BeforeEach(func() {
		cleanupDatabase()
	})

	AfterEach(func() {
		// cleanupDatabase()
	})

	It("create migration", func() {
		err := gomigrator.DoCreate(testConfig, "api1")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoCreate(testConfig, "api2")
		Expect(err).NotTo(HaveOccurred())

		// Check SQL migration
		err = gomigrator.DoCreate(testConfig, "mig1")
		Expect(err).NotTo(HaveOccurred())

		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(3))

		Expect(hasMigrations[0].Mname).Should(Equal("api1"))
		Expect(hasMigrations[0].Mtype).Should(Equal("go"))

		Expect(hasMigrations[1].Mname).Should(Equal("api2"))
		Expect(hasMigrations[1].Mtype).Should(Equal("go"))

		Expect(hasMigrations[2].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[2].Mtype).Should(Equal("sql"))
	})

	It("status", func() {
		// set up
		err := gomigrator.DoCreate(testConfig, "api1")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoCreate(testConfig, "api2")
		Expect(err).NotTo(HaveOccurred())

		// Check SQL migration
		err = gomigrator.DoCreate(testConfig, "mig1")
		Expect(err).NotTo(HaveOccurred())

		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(3))

		// run test
		migrations, err := gomigrator.Status(testConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(migrations)).Should(Equal(3))

		Expect(migrations[0].Name).Should(Equal("api1"))
		Expect(migrations[0].Type).Should(Equal("go"))
		Expect(migrations[0].Status).Should(Equal("new"))
		Expect(migrations[0].Applied).Should(Equal(false))

		Expect(migrations[1].Name).Should(Equal("api2"))
		Expect(migrations[1].Type).Should(Equal("go"))
		Expect(migrations[1].Status).Should(Equal("new"))
		Expect(migrations[1].Applied).Should(Equal(false))

		Expect(migrations[2].Name).Should(Equal("mig1"))
		Expect(migrations[2].Type).Should(Equal("sql"))
		Expect(migrations[2].Status).Should(Equal("new"))
		Expect(migrations[2].Applied).Should(Equal(false))
	})

	It("dbversion", func() {
		err := gomigrator.DoCreate(testConfig, "api1")
		Expect(err).NotTo(HaveOccurred())

		_, err = gomigrator.DBVersion(testConfig)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(sql.ErrNoRows))

		err = gomigrator.DoUp(testConfig)
		Expect(err).NotTo(HaveOccurred())

		version, err := gomigrator.DBVersion(testConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(version.Version).Should(Equal(1))
		Expect(version.MigrationName).Should(Equal("api1"))
	})
	It("up", func() {
		err := gomigrator.DoCreate(testConfig, "api1")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoUp(testConfig)
		Expect(err).NotTo(HaveOccurred())

		// Check that mig1 is in migrator table and status is applied
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(1))
		Expect(hasMigrations[0].Mname).Should(Equal("api1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))

		value := selectValue("apitest1", "i")
		Expect(value).Should(Equal(two))
	})

	It("up2", func() {
		err := gomigrator.DoCreate(testConfig, "api1")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoCreate(testConfig, "api2")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoCreate(testConfig, "mig1")
		Expect(err).NotTo(HaveOccurred())

		err = gomigrator.DoUp(testConfig)
		Expect(err).NotTo(HaveOccurred())

		// Check that mig1 is in migrator table and status is applied
		checkMigrationsStatus([]string{"applied", "applied", "applied"})

		value := selectValue("apitest1", "i")
		Expect(value).Should(Equal(two))

		value2 := selectValue("apitest2", "key")
		Expect(value2).Should(Equal("two"))

		value = selectValue("test1", "i")
		Expect(value).Should(Equal(one))
	})

	It("down", func() {
		setUpMigrations()
		// down test
		err = gomigrator.DoDown(testConfig)
		Expect(err).NotTo(HaveOccurred())

		checkMigrationsStatus([]string{"applied", "applied", "new"})

		tables := getTables()
		Expect(tables).To(ContainElement("apitest1"))
		Expect(tables).To(ContainElement("apitest2"))
		Expect(tables).NotTo(ContainElement("test1"))
	})

	It("redo", func() {
		setUpMigrations()

		// redo test
		err = gomigrator.DoRedo(testConfig)
		Expect(err).NotTo(HaveOccurred())

		// Check that mig1 and mig2 are in migrator table and status is new
		hasMigrations := getMigrationRecords()
		Expect(len(hasMigrations)).Should(Equal(3))
		Expect(hasMigrations[0].Mname).Should(Equal("api1"))
		Expect(hasMigrations[0].Mstatus).Should(Equal("applied"))
		Expect(hasMigrations[1].Mname).Should(Equal("api2"))
		Expect(hasMigrations[1].Mstatus).Should(Equal("applied"))
		Expect(hasMigrations[2].Mname).Should(Equal("mig1"))
		Expect(hasMigrations[2].Mstatus).Should(Equal("applied"))

		tables := getTables()
		Expect(tables).To(ContainElement("apitest1"))
		Expect(tables).To(ContainElement("apitest2"))
		Expect(tables).To(ContainElement("test1"))
	})
})
