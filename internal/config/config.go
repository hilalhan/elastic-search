package config

import (
	"fmt"
	"strings"

	"github.com/ory/viper"
)

// ----- Environment types -----
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// ----- Server configuration -----
type ServerConfig struct {
	Address         string `mapstructure:"SERVER_ADDRESS"`
	ReadTimeoutSec  int    `mapstructure:"SERVER_READ_TIMEOUT_SEC"`
	WriteTimeoutSec int    `mapstructure:"SERVER_WRITE_TIMEOUT_SEC"`
	IdleTimeoutSec  int    `mapstructure:"SERVER_IDLE_TIMEOUT_SEC"`
}

// ----- Elasticsearch configuration -----
type ElasticsearchConfig struct {
	Addresses  []string `mapstructure:"ELASTICSEARCH_ADDRESSES"`
	Username   string   `mapstructure:"ELASTICSEARCH_USERNAME"`
	Password   string   `mapstructure:"ELASTICSEARCH_PASSWORD"`
	Index      string   `mapstructure:"ELASTICSEARCH_INDEX"`
	TimeoutSec int      `mapstructure:"ELASTICSEARCH_TIMEOUT_SEC"`
}

// ----- Main configuration struct -----
type Config struct {
	Environment   Environment `mapstructure:"ENVIRONMENT"`
	Server        ServerConfig
	Elasticsearch ElasticsearchConfig
}

// Load loads the configuration from .env file
func Load() (*Config, error) {
	v := viper.New()

	// Set up Viper for .env file
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	// Enable environment variables
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create config with default values
	cfg := Config{
		Environment: EnvDevelopment,
		Server: ServerConfig{
			Address:         ":8080",
			ReadTimeoutSec:  30,
			WriteTimeoutSec: 30,
			IdleTimeoutSec:  60,
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses:  []string{"http://localhost:9200"},
			Index:      "documents",
			TimeoutSec: 10,
		},
	}

	if env := v.GetString("ENVIRONMENT"); env != "" {
		cfg.Environment = Environment(env)
	}

	if serverAddress := v.GetString("SERVER_ADDRESS"); serverAddress != "" {
		cfg.Server.Address = serverAddress
	}

	if serverIdleTimeout := v.GetInt("SERVER_IDLE_TIMEOUT_SEC"); serverIdleTimeout != 0 {
		cfg.Server.IdleTimeoutSec = serverIdleTimeout
	}

	if serverWriteTimeout := v.GetInt("SERVER_WRITE_TIMEOUT_SEC"); serverWriteTimeout != 0 {
		cfg.Server.WriteTimeoutSec = serverWriteTimeout
	}

	if serverReadTimeout := v.GetInt("SERVER_READ_TIMEOUT_SEC"); serverReadTimeout != 0 {
		cfg.Server.ReadTimeoutSec = serverReadTimeout
	}

	if serverReadTimeout := v.GetInt("SERVER_READ_TIMEOUT_SEC"); serverReadTimeout != 0 {
		cfg.Server.ReadTimeoutSec = serverReadTimeout
	}

	if esAddresses := v.GetString("ELASTICSEARCH_ADDRESSES"); esAddresses != "" {
		cfg.Elasticsearch.Addresses = strings.Split(esAddresses, ",")
	}

	if esIndex := v.GetString("ELASTICSEARCH_INDEX"); esIndex != "" {
		cfg.Elasticsearch.Index = esIndex
	}

	if esTimeout := v.GetInt("ELASTICSEARCH_TIMEOUT_SEC"); esTimeout != 0 {
		cfg.Elasticsearch.TimeoutSec = esTimeout
	}

	if esUsername := v.GetString("ELASTICSEARCH_USERNAME"); esUsername != "" {
		cfg.Elasticsearch.Username = esUsername
	}

	if esPassword := v.GetString("ELASTICSEARCH_PASSWORD"); esPassword != "" {
		cfg.Elasticsearch.Password = esPassword
	}

	return &cfg, nil
}

// Helper methods to access configuration values
func GetBool(key string) bool {
	return viper.GetBool(key)
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetIntSlice(key string) []int {
	return viper.GetIntSlice(key)
}

func GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}
