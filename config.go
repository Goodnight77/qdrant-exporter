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
		// It's okay if .env doesn't exist
	}

	viper.SetDefault("qdrant_url", "http://localhost:6333")
	viper.SetDefault("listen_address", "0.0.0.0:9090")
	viper.SetDefault("scrape_interval", 15*time.Second)
	viper.SetDefault("log_level", "info")

	// Map Environment variables to config struct
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	// Explicitly bind for cases where names might not match perfectly
	viper.BindEnv("qdrant_url", "QDRANT_URL")
	viper.BindEnv("qdrant_api_key", "QDRANT_API_KEY")
	viper.BindEnv("listen_address", "LISTEN_ADDRESS")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
