package gomigrator

import (
	"testing"

	"github.com/dkovalev1/gomigrator/config" //nolint
	"github.com/stretchr/testify/require"    //nolint
)

func TestRegistry(t *testing.T) {
	testConfig := config.Config{}

	t.Run("Create", func(t *testing.T) {
		err := DoCreate(testConfig)
		require.Error(t, err)
	})
}
