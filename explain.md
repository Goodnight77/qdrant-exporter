# Understanding the Qdrant Exporter (Go Code Explanation)

This document explains the codebase for the **Qdrant Exporter**. Since you are new to Go, we will break down the structure, key concepts, and how the different files work together to expose Qdrant metrics to Prometheus.

## Project Overview

**Goal:** This application fetches statistics (like vector counts, shard status) from a running Qdrant instance and exposes them in a format that Prometheus can "scrape" (collect).

**Key Go Concepts Used:**
- **Structs**: Used to define data structures (like configuration, API responses).
- **Interfaces**: specifically `prometheus.Collector` to integrate with the Prometheus library.
- **HTTP Server**: To serve the metrics on a specific port.
- **Goroutines**: Lightweight threads (used for the background ticker and serving requests).
- **Channels**: Used in the collector to send metrics back to Prometheus.

---

## File-by-File Breakdown

### 1. `main.go` - The Entry Point
This is where the program starts. Its job is to "wire everything together."

- **What it does:**
    1.  **Parses Flags & Config**: It reads command-line arguments (like `--qdrant-url`) and configuration files/env vars.
    2.  **Sets up Logging**: Initializes the `zap` logger for structured logs.
    3.  **Initializes Components**: Creates instances of the `QdrantClient` and `QdrantCollector`.
    4.  **Registers the Collector**: Tells Prometheus "Hey, use this collector to get metrics."
    5.  **Starts the HTTP Server**: Listens on a port (default `:9090`) to serve endpoints like `/metrics` and `/health`.

- **Key Code Snippet:**
  ```go
  // Connects the collector to Prometheus
  prometheus.MustRegister(collector)

  // Starts the web server
  http.Handle("/metrics", promhttp.Handler())
  http.ListenAndServe(cfg.ListenAddress, nil)
  ```

### 2. `config.go` - Configuration Management
This file handles how the application gets its settings.

- **What it does:**
    - Defines a `Config` struct that mirrors the settings we need (URL, API Key, etc.).
    - Uses `viper`, a popular configuration library, to read from:
        1. Defaults (hardcoded fallbacks).
        2. `.env` file (for local development).
        3. Environment variables (great for Docker/Kubernetes).
    - Returns a filled `Config` object to usage in `main.go`.

### 3. `client.go` - The Qdrant API Client
This file is responsible for talking to the Qdrant database. It isolates the "how do I get data" logic.

- **What it does:**
    - Defines structs (like `CollectionsResponse`, `CollectionInfo`) that match the **JSON** returned by Qdrant's API.
    - Uses Go's standard `net/http` library to make `GET` requests to Qdrant.
    - `GetCollections()`: Asks Qdrant for a list of all collections.
    - `GetCollectionInfo()`: Asks for details (vector count, status) of a specific collection.
    - `GetCollectionClusterInfo()`: Asks for shard details (useful if Qdrant is running in a cluster).

- **Why it matters:**  
  By separating this logic, the rest of the app doesn't need to know *how* to call Qdrant, just that it can get the data.

### 4. `collector.go` - The Prometheus Collector
This is the core logic for the exporter. It implements the **Prometheus Collector Interface**.

- **What it does:**
    - **`Describe` method**: Tells Prometheus what metrics are available (e.g., "I provide `qdrant_collection_vectors_total`").
    - **`Collect` method**: This is called every time Prometheus scrapes the exporter.
        1. It uses `client.go` to fetch fresh data from Qdrant.
        2. It calculates metrics (e.g., counting shards).
        3. It sends these values into a **channel** (`ch <- metric`).
    
- **Key Metrics Explained:**
    - `qdrant_collection_vectors_total`: How many vectors you have.
    - `qdrant_collection_points_total`: Total points (vectors + payload).
    - `qdrant_collection_shards_total`: Health status of your data shards.

- **Concurrency Note:**
  The `Collect` function is designed to be thread-safe. It performs the API calls *on demand* when Prometheus asks for data.

### 5. `collector_test.go` - Unit Testing
This file ensures the collector works correctly without needing a real Qdrant instance.

- **What it does:**
    - **Mocks Qdrant**: It creates a fake HTTP server (`httptest.NewServer`) that pretends to be Qdrant and returns dummy JSON.
    - **Runs the Collector**: It manually calls the `Collect` method.
    - **Verifies Output**: It checks if the collector actually produced the expected metrics (like checking if "vectors_count" was read correctly).

## How Data Flows

1. **Prometheus** (external) sends a HTTP GET request to `http://localhost:9090/metrics`.
2. The `promhttp.Handler` in `main.go` receives this request.
3. It calls `collector.Collect()`.
4. `collector.Collect()` calls `client.GetCollections()`, then loops through them to get details.
5. The `client` makes HTTP requests to **Qdrant**.
6. Qdrant returns JSON data.
7. The `client` parses JSON into Go structs.
8. The `collector` converts these structs into **Prometheus Metrics**.
9. The metrics are returned to Prometheus as the HTTP response.

## Next Steps for Learning
- Try changing a metric name in `collector.go` and see it update in the output.
- Look at `client.go` structs and compare them to Qdrant's API documentation.
- Run `go test ./...` to see the tests in action.
