package internal

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require" //nolint
)

func testUp(_ *sql.Tx) error {
	return errors.New("42")
}

func testDown(_ *sql.Tx) error {
	return errors.New("43")
}

func cleanupRegistry() {
	r := Registry.(*RegistryImpl)
	clear(r.records)
}

func TestRegistry(t *testing.T) {
	t.Run("Registry basic", func(t *testing.T) {
		cleanupRegistry()
		require.NotNil(t, Registry)

		Registry.Register("mig2go", testUp, testDown)

		r := Registry.(*RegistryImpl)

		require.NotNil(t, r)
		require.Equal(t, 1, len(r.records))

		Registry.Register("mig2go2", testUp, testDown)
		require.Equal(t, len(r.records), 2)

		require.True(t, Registry.Check("mig2go2"))

		m := Registry.Get("mig2go2")
		require.NotNil(t, m)
		require.Equal(t, "mig2go2", m.Name)
		require.Equal(t, "43", m.Down(nil).Error())
		require.Equal(t, "42", m.Up(nil).Error())
	})
}
