package gomigrator

import (
	"testing"

	"github.com/dkovalev1/gomigrator/config"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {

	test_config := config.Config{}

	t.Run("Create", func(t *testing.T) {
		err := DoCreate(test_config, make([]string, 0))
		require.Error(t, err)
	})
}
