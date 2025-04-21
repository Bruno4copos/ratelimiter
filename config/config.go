package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	MaxRequestsPerSecondIP    int           `mapstructure:"MAX_REQUESTS_PER_SECOND_IP"`
	MaxRequestsPerSecondToken int           `mapstructure:"MAX_REQUESTS_PER_SECOND_TOKEN"`
	WebServerPort             int           `mapstructure:"WEB_SERVER_PORT"`
	BlockDurationIP           time.Duration `mapstructure:"BLOCK_DURATION_IP"`
	BlockDurationToken        time.Duration `mapstructure:"BLOCK_DURATION_TOKEN"`
	RedisAddress              string        `mapstructure:"REDIS_ADDRESS"`
	RedisPassword             string        `mapstructure:"REDIS_PASSWORD"`
	Tokens                    string        `mapstructure:"TOKENS"`
	TokensMap                 map[string]Token
}

type Token struct {
	RateLimit    int
	RateInterval int
}

func LoadConfig(path string) (*Config, error) {

	var (
		config = &Config{} // environment variables saved in the .env file
		err    error
	)
	viper.SetDefault("MAX_REQUESTS_PER_SECOND_IP", 5)
	viper.SetDefault("MAX_REQUESTS_PER_SECOND_TOKEN", 100)
	viper.SetDefault("BLOCK_DURATION_IP", "60s")
	viper.SetDefault("BLOCK_DURATION_TOKEN", "60s")
	viper.SetDefault("REDIS_ADDRESS", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// .env file not found, using default values
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	config.TokensMap, err = parseTokens(config.Tokens)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func parseTokens(tokensString string) (map[string]Token, error) {

	var (
		tokens                  map[string]Token
		pairs, kv, values       []string
		pair                    string
		rateLimit, rateInterval int
		err                     error
	)
	tokens = make(map[string]Token)
	fmt.Printf("TOKEN: %v\n", tokensString)
	pairs = strings.Split(tokensString, ",")
	for _, pair = range pairs {
		kv = strings.Split(pair, ":")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid token format: %s", pair)
		}
		values = strings.Split(kv[1], "/")
		if len(values) != 2 {
			return nil, fmt.Errorf("invalid token values format: %s", kv[1])
		}
		rateLimit, err = strconv.Atoi(values[0])
		if err != nil {
			return nil, fmt.Errorf("invalid rate limit value: %s", values[0])
		}
		rateInterval, err = strconv.Atoi(values[1])
		if err != nil {
			return nil, fmt.Errorf("invalid rate interval value: %s", values[1])
		}
		tokens[kv[0]] = Token{
			RateLimit:    rateLimit,
			RateInterval: rateInterval,
		}
	}
	return tokens, nil
}
