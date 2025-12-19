package main

import (
	"encoding/json" // json parsing responses from Qdrant
	"fmt"
	"io" // reading HTTP response bodies
	"net/http" // making http requests & creating server
	"time" // for sleep/delay between scrapes

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

	// for /metrics endpoint (standard Prometheus format)
	http.Handle("/metrics", promhttp.Handler())

	// start background goroutine to scrape Qdrant and update metrics
	go scrapeQdrant() // runs forever

	fmt.Println("server starting on http://localhost:9999")
	fmt.Println("metrics available at http://localhost:9999/metrics")
	fmt.Println("Qdrant should be running on http://localhost:6333")
	fmt.Println("\n--- press Ctrl+C to stop ---\n")

	err := http.ListenAndServe(":9999", nil) // start http server
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

// scrapeQdrant runs in background to fetch data from Qdrant and update metrics
func scrapeQdrant() {
	qdrantURL := "http://localhost:6333"

	for { // inf loop
		// get list of all collections
		resp, err := http.Get(qdrantURL + "/collections")
		if err != nil {
			fmt.Printf("# Error connecting to Qdrant: %v\n", err)
			qdrantUp.Set(0) // set to 0 if connection fails
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("# Error reading response: %v\n", err)
			qdrantUp.Set(0)
			continue
		}

		var collectionsResp CollectionsResponse
		if err := json.Unmarshal(body, &collectionsResp); err != nil {
			fmt.Printf("# Error parsing JSON: %v\n", err)
			qdrantUp.Set(0)
			continue
		}

		// connection successful, set qdrant_up to 1
		qdrantUp.Set(1)

		// For each collection, get detailed info
		// logged now later exposed as metrics
		for _, col := range collectionsResp.Result.Collections {
			collectionName := col.Name

			// Call Qdrant API for this specific collection
			infoURL := fmt.Sprintf("%s/collections/%s", qdrantURL, collectionName)
			infoResp, err := http.Get(infoURL)

			if err != nil {
				// if we can't get info for this collection, skip it
				fmt.Printf("# Error getting info for %s: %v\n", collectionName, err)
				continue // skip to next coll
			}

			infoBody, _ := io.ReadAll(infoResp.Body) // read first
			infoResp.Body.Close() // then close

			var info CollectionInfoResponse
			if err := json.Unmarshal(infoBody, &info); err != nil {
				fmt.Printf("# Error parsing info for %s: %v\n", collectionName, err)
				continue
			}

			// log collection info
			fmt.Printf("collection=%s points=%d vectors=%d indexed=%d segments=%d status=%s\n",
				collectionName,
				info.Result.PointsCount,
				info.Result.VectorsCount,
				info.Result.IndexedVectors,
				info.Result.SegmentsCount,
				info.Result.Status)
		}

		// empty line between scrapes
		fmt.Println()

		// wait before next scrape (every 10)
		time.Sleep(10 * time.Second)
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
