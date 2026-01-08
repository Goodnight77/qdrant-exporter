package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQdrantClient_NewFeatures(t *testing.T) {
	// Mock Qdrant API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/collections/test_conn":
			// Return vector config
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"result": {
					"config": {
						"params": {
							"vectors": {
								"size": 768,
								"distance": "Cosine"
							}
						}
					}
				},
				"status": "ok"
			}`)
		case "/collections/test_conn/points/scroll":
			// Return scrolled points
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"result": {
					"points": [
						{
							"id": "point-1",
							"payload": {"city": "Berlin", "country": "Germany"}
						},
						{
							"id": 2,
							"payload": {"city": "London", "country": "UK"}
						}
					]
				},
				"status": "ok"
			}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewQdrantClient(server.URL, "test-key")

	// Test GetCollectionVectorSize
	t.Run("GetCollectionVectorSize", func(t *testing.T) {
		size, err := client.GetCollectionVectorSize("test_conn")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if size != 768 {
			t.Errorf("Expected size 768, got %d", size)
		}
	})

	// Test ScrollPoints
	t.Run("ScrollPoints", func(t *testing.T) {
		points, err := client.ScrollPoints("test_conn")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(points) != 2 {
			t.Errorf("Expected 2 points, got %d", len(points))
		}
		
		if points[0].ID != "point-1" {
			t.Errorf("Expected first point ID 'point-1', got %v", points[0].ID)
		}
		// String comparison for JSON payload can be flaky due to key ordering, but let's check content exists
		if points[0].Payload == "" {
			t.Error("Expected payload for point 1")
		}
		
		if points[1].ID != "2" {
			t.Errorf("Expected second point ID '2', got %v", points[1].ID)
		}
	})
}
