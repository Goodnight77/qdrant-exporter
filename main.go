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
	// defer logger.Sync() // flushes buffer, if any. ignored because often fails on stdout/stderr
	defer func() { _ = logger.Sync() }()

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
			if _, err := fmt.Fprintf(w, `{"status": "degraded", "last_success": "%s"}`, lastSuccess.Format(time.RFC3339)); err != nil {
				logger.Error("Failed to write health response", zap.Error(err))
			}
			return
		}
		if _, err := fmt.Fprintf(w, `{"status": "ok", "last_success": "%s"}`, lastSuccess.Format(time.RFC3339)); err != nil {
			logger.Error("Failed to write health response", zap.Error(err))
		}
	})
	// Inspect Handler to view collection details (vector size, points sample)
	http.HandleFunc("/inspect", func(w http.ResponseWriter, r *http.Request) {
		collectionName := r.URL.Query().Get("collection")
		if collectionName == "" {
			http.Error(w, "Missing collection parameter", http.StatusBadRequest)
			return
		}

		vectorSize, err := qdrantClient.GetCollectionVectorSize(collectionName)
		if err != nil {
			logger.Error("Failed to get vector size", zap.Error(err))
			http.Error(w, fmt.Sprintf("Failed to get vector size: %v", err), http.StatusInternalServerError)
			return
		}

		points, err := qdrantClient.ScrollPoints(collectionName)
		if err != nil {
			logger.Error("Failed to scroll points", zap.Error(err))
			http.Error(w, fmt.Sprintf("Failed to scroll points: %v", err), http.StatusInternalServerError)
			return
		}

		html := fmt.Sprintf(`<html>
			<head>
				<title>Inspect %s</title>
				<style>
					table { border-collapse: collapse; width: 100%%; }
					th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
					tr:nth-child(even) { background-color: #f2f2f2; }
					th { background-color: #4CAF50; color: white; }
				</style>
			</head>
			<body>
			<h1>Collection: %s</h1>
			<p><a href="/">Back to Home</a></p>
			<h3>Metadata</h3>
			<ul>
				<li><strong>Vector Size:</strong> %d</li>
			</ul>
			<h3>Sample Points (Top 10)</h3>
			<table>
				<tr>
					<th>ID</th>
					<th>Payload (Document)</th>
				</tr>`, collectionName, collectionName, vectorSize)

		for _, p := range points {
			html += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", p.ID, p.Payload)
		}
		html += `</table></body></html>`

		if _, err := w.Write([]byte(html)); err != nil {
			logger.Error("Failed to write inspect response", zap.Error(err))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		collections, err := qdrantClient.GetCollections()
		collectionsHtml := "<ul>"
		if err != nil {
			collectionsHtml += fmt.Sprintf("<li>Error fetching collections: %v</li>", err)
		} else {
			for _, c := range collections {
				collectionsHtml += fmt.Sprintf(`<li><a href="/inspect?collection=%s">%s</a></li>`, c, c)
			}
		}
		collectionsHtml += "</ul>"

		if _, err := w.Write([]byte(fmt.Sprintf(`<html>
			<head><title>Qdrant Exporter</title></head>
			<body>
			<h1>Qdrant Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			<p><a href="/health">Health</a></p>
			<h3>Collections</h3>
			%s
			</body>
			</html>`, collectionsHtml))); err != nil {
			logger.Error("Failed to write root response", zap.Error(err))
		}
	})

	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		logger.Fatal("Error starting HTTP server", zap.Error(err))
	}
}
