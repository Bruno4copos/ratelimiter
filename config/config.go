package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	MaxRequestsPerSecondIP    int
	MaxRequestsPerSecondToken int
	BlockDurationIP           time.Duration
	BlockDurationToken        time.Duration
	RedisAddress              string
	RedisPassword             string
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("MAX_REQUESTS_PER_SECOND_IP", 5)
	viper.SetDefault("MAX_REQUESTS_PER_SECOND_TOKEN", 100)
	viper.SetDefault("BLOCK_DURATION_IP", "5m")
	viper.SetDefault("BLOCK_DURATION_TOKEN", "1h")
	viper.SetDefault("REDIS_ADDRESS", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// .env file not found, using default values
	}

	config := &Config{
		MaxRequestsPerSecondIP:    viper.GetInt("MAX_REQUESTS_PER_SECOND_IP"),
		MaxRequestsPerSecondToken: viper.GetInt("MAX_REQUESTS_PER_SECOND_TOKEN"),
		RedisAddress:              viper.GetString("REDIS_ADDRESS"),
		RedisPassword:             viper.GetString("REDIS_PASSWORD"),
	}

	blockDurationIPStr := viper.GetString("BLOCK_DURATION_IP")
	config.BlockDurationIP, err = time.ParseDuration(blockDurationIPStr)
	if err != nil {
		return nil, err
	}

	blockDurationTokenStr := viper.GetString("BLOCK_DURATION_TOKEN")
	config.BlockDurationToken, err = time.ParseDuration(blockDurationTokenStr)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
