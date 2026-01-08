package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// CLI Flags
	qdrantURL := flag.String("qdrant-url", "", "Qdrant API URL")
	qdrantAPIKey := flag.String("qdrant-api-key", "", "Qdrant API Key")
	listenAddr := flag.String("listen-address", "", "Address to listen on for HTTP requests")
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Load Config
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	// Override config with flags if provided
	if *qdrantURL != "" {
		cfg.QdrantURL = *qdrantURL
	}
	if *qdrantAPIKey != "" {
		cfg.QdrantAPIKey = *qdrantAPIKey
	}
	if *listenAddr != "" {
		cfg.ListenAddress = *listenAddr
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// Setup Logger
	loggerConfig := zap.NewProductionConfig()
	if cfg.LogLevel == "debug" {
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, _ := loggerConfig.Build()
	defer logger.Sync()

	logger.Info("Starting Qdrant exporter",
		zap.String("listen_address", cfg.ListenAddress),
		zap.String("qdrant_url", cfg.QdrantURL),
		zap.String("log_level", cfg.LogLevel),
		zap.Bool("with_api_key", cfg.QdrantAPIKey != ""),
	)

	qdrantClient := NewQdrantClient(cfg.QdrantURL, cfg.QdrantAPIKey)
	collector := NewQdrantCollector(qdrantClient, logger)
	prometheus.MustRegister(collector)

	// Periodic logging (as requested in Task 2, but now using logger)
	go func() {
		ticker := time.NewTicker(cfg.ScrapeInterval)
		defer ticker.Stop()

		for {
			collections, err := qdrantClient.GetCollections()
			if err != nil {
				logger.Error("Failed to fetch collections", zap.Error(err))
			} else {
				logger.Info("Found collections", zap.Int("count", len(collections)), zap.Strings("names", collections))
			}
			<-ticker.C
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		lastSuccess := collector.GetLastSuccessTime()
		if time.Since(lastSuccess) > cfg.ScrapeInterval*2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "degraded", "last_success": "%s"}`, lastSuccess.Format(time.RFC3339))
			return
		}
		fmt.Fprintf(w, `{"status": "ok", "last_success": "%s"}`, lastSuccess.Format(time.RFC3339))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Qdrant Exporter</title></head>
			<body>
			<h1>Qdrant Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			<p><a href="/health">Health</a></p>
			</body>
			</html>`))
	})

	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Fatal("Error starting HTTP server", zap.Error(err))
	}
}
