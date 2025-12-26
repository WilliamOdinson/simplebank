package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application
// loaded from environment variables
type Config struct {
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// environment variables
	viper.BindEnv("DB_SOURCE")
	viper.BindEnv("SERVER_ADDRESS")
	viper.BindEnv("TOKEN_SYMMETRIC_KEY")
	viper.BindEnv("ACCESS_TOKEN_DURATION")

	// Try to read config file (if it exists)
	viper.ReadInConfig()

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	if config.DBSource == "" {
		panic("DB_SOURCE is required")
	}
	if config.ServerAddress == "" {
		panic("SERVER_ADDRESS is required")
	}
	if config.TokenSymmetricKey == "" {
		panic("TOKEN_SYMMETRIC_KEY is required")
	}
	if config.AccessTokenDuration == 0 {
		panic("ACCESS_TOKEN_DURATION is required")
	}

	return
}
