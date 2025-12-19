package main

import (
	"fmt"
	"net/http" // creating server

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// response from /collections endpoint 
type CollectionsResponse struct {
	Result struct {
		Collections []struct {
			Name string `json:"name"`
		} `json:"collections"`
	} `json:"result"`
}

// response from /collections/{name} endpoint + details 
type CollectionInfoResponse struct { 
	Result struct {
		PointsCount    uint64 `json:"points_count"` // nbre of data pts 
		VectorsCount   uint64 `json:"vectors_count"` // nbre of vectors
		IndexedVectors uint64 `json:"indexed_vectors_count"` // nbre of indexed vectors
		SegmentsCount  uint64 `json:"segments_count"` // nbre of internal storage segments
		Status         string `json:"status"` // color status of collection (green/yellow/red)
	} `json:"result"`
}

// gauge metric to show if Qdrant is reachable
var qdrantUp = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "qdrant_up",
	Help: "Whether Qdrant is reachable",
})

func main() {
	// register the gauge metric with Prometheus
	prometheus.MustRegister(qdrantUp)

	// create qdrant client
	client := NewQdrantClient("http://localhost:6333")

	// create and register the collector
	collector := NewQdrantCollector(client)
	prometheus.MustRegister(collector)

	// for /metrics endpoint (standard Prometheus format)
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("server starting on http://localhost:9999")
	fmt.Println("metrics available at http://localhost:9999/metrics")
	fmt.Println("Qdrant should be running on http://localhost:6333")
	fmt.Println("\n--- press Ctrl+C to stop ---\n")

	err := http.ListenAndServe(":9999", nil) // start http server
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

// convert Qdrant status string to a number for Prometheus
// Prometheus metrics work best with numbers
func statusToNumber(status string) int {
	switch status { // switch is like if-else for multiple conditions
	case "green":
		return 1 // all good
	case "yellow":
		return 2 // warning
	case "red":
		return 3 // error
	default:
		return 0 // unknown
	}
}
