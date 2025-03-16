package internal

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func test_up(tx *sql.Tx) error {
	return errors.New("42")
}

func test_down(tc *sql.Tx) error {
	return errors.New("43")
}

func TestRegistry(t *testing.T) {

	t.Run("Registry basic", func(t *testing.T) {
		require.NotNil(t, Registry)

		Registry.Register("mig2go", test_up, test_down)

		r := Registry.(*RegistryImpl)

		require.NotNil(t, r)
		require.Equal(t, len(r.records), 1)

		Registry.Register("mig2go2", test_up, test_down)
		require.Equal(t, len(r.records), 2)

		require.True(t, Registry.Check("mig2go2"))

		m := Registry.Get("mig2go2")
		require.NotNil(t, m)
		require.Equal(t, "mig2go2", m.Name)
		require.Equal(t, "43", m.Down(nil).Error())
		require.Equal(t, "42", m.Up(nil).Error())
	})
}
