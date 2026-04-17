package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// SovereignConfig holds all configuration for the system.
type SovereignConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Inference InferenceConfig `mapstructure:"inference"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Production bool   `mapstructure:"production"`
}

type StorageConfig struct {
	DataDir string `mapstructure:"data_dir"`
}

type InferenceConfig struct {
	Device string `mapstructure:"device"` // cpu, cuda, vulkan
}

// Load initializes configuration from files, environment variables, and defaults.
func Load(path string) (*SovereignConfig, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.env", "development")
	v.SetDefault("logging.level", "info")
	v.SetDefault("storage.data_dir", "./data")
	v.SetDefault("inference.device", "cpu")

	// Environment variables
	v.SetEnvPrefix("SOVEREIGN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg SovereignConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
