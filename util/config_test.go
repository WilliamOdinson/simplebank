package util

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("..")
	require.NoError(t, err)
	require.NotEmpty(t, config.DBSource)
	require.NotEmpty(t, config.ServerAddress)
	require.NotEmpty(t, config.TokenSymmetricKey)
	require.NotZero(t, config.AccessTokenDuration)
}

func TestLoadConfigNotFound(t *testing.T) {
	viper.Reset()
	require.Panics(t, func() {
		LoadConfig("/nonexistent/path")
	})
}
