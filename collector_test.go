package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func TestCollector_Collect(t *testing.T) {
	// Mock Qdrant API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/collections":
			fmt.Fprint(w, `{"result": {"collections": [{"name": "test_collection"}]}, "status": "ok"}`)
		case "/collections/test_collection":
			fmt.Fprint(w, `{"result": {"vectors_count": 100, "points_count": 100, "status": "green"}, "status": "ok"}`)
		case "/collections/test_collection/cluster":
			fmt.Fprint(w, `{"result": {"peer_id": 1, "shards": [{"shard_id": 0, "status": "active"}]}, "status": "ok"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewQdrantClient(server.URL, "")
	collector := NewQdrantCollector(client, logger)

	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metricsFound := 0
	for range ch {
		metricsFound++
	}

	if metricsFound < 3 {
		t.Errorf("Expected at least 3 metrics, found %d", metricsFound)
	}
}
