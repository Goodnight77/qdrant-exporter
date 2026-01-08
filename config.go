package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	QdrantURL      string        `mapstructure:"qdrant_url"`
	QdrantAPIKey   string        `mapstructure:"qdrant_api_key"`
	ListenAddress  string        `mapstructure:"listen_address"`
	ScrapeInterval time.Duration `mapstructure:"scrape_interval"`
	LogLevel       string        `mapstructure:"log_level"`
}

func LoadConfig() (*Config, error) {
	// Try to load .env file (if running locally)
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	viper.SetDefault("qdrant_url", "http://localhost:6333")
	viper.SetDefault("listen_address", "0.0.0.0:9090")
	viper.SetDefault("scrape_interval", 15*time.Second)
	viper.SetDefault("log_level", "info")

	// Map Environment variables to config struct
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	// Explicitly bind for cases where names might not match perfectly
	if err := viper.BindEnv("qdrant_url", "QDRANT_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind qdrant_url env: %w", err)
	}
	if err := viper.BindEnv("qdrant_api_key", "QDRANT_API_KEY"); err != nil {
		return nil, fmt.Errorf("failed to bind qdrant_api_key env: %w", err)
	}
	if err := viper.BindEnv("listen_address", "LISTEN_ADDRESS"); err != nil {
		return nil, fmt.Errorf("failed to bind listen_address env: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
