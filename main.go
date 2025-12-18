package main

import (
	"encoding/json" // json parsing responses from Qdrant
	"fmt"
	"io" // reading HTTP response bodies
	"net/http" // making http requests & creating server 
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

func main() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { // create /metrics endpoint 
		qdrantURL := "http://localhost:6333"

		// STEP 1: Get list of all collections
		resp, err := http.Get(qdrantURL + "/collections") // get all colls call this endpoint 6333/collections
		if err != nil { 
			fmt.Fprintf(w, "# Error connecting to Qdrant: %v\n", err) // write directly to http resp (w is the http.responsewriter) # is a comment in Prometheus metrics
			fmt.Fprintf(w, "qdrant_up 0\n") // if no connection we set this metric to 0 (down)
			return // exit 
		}
		defer resp.Body.Close() // run this after fun finish even with error/panic 

		body, err := io.ReadAll(resp.Body) // read resp body into byte slice []byte  return ([]byte, error)
		if err != nil {
			fmt.Fprintf(w, "# Error reading response: %v\n", err)
			return
		}
// parse JSON response into our struct
		var collectionsResp CollectionsResponse // var of type collresp 
		if err := json.Unmarshal(body, &collectionsResp); err != nil { // parse json bytes into struct , &collectionsresp is pointer to var (needed so it modify variable )
			fmt.Fprintf(w, "# Error parsing JSON: %v\n", err)
			return
		}

		// output basic metrics
		fmt.Fprintln(w, "# HELP qdrant_up Whether Qdrant is reachable")
		fmt.Fprintln(w, "# TYPE qdrant_up gauge")
		fmt.Fprintf(w, "qdrant_up 1\n")

		// STEP 3: For EACH collection, get detailed info
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
			defer infoResp.Body.Close()

			infoBody, _ := io.ReadAll(infoResp.Body) // ignore err handling 

			var info CollectionInfoResponse
			if err := json.Unmarshal(infoBody, &info); err != nil {
				fmt.Printf("# Error parsing info for %s: %v\n", collectionName, err)
				continue
			}

			// output metrics with LABELS
			// labels are like key-value pairs that identify the collection

			// Points count
			fmt.Fprintf(w, "qdrant_collection_points{collection=\"%s\"} %d\n",
				collectionName, info.Result.PointsCount)

			// Vectors count 
			fmt.Fprintf(w, "qdrant_collection_vectors{collection=\"%s\"} %d\n",
				collectionName, info.Result.VectorsCount)

			// Indexed vectors 
			fmt.Fprintf(w, "qdrant_collection_indexed_vectors{collection=\"%s\"} %d\n",
				collectionName, info.Result.IndexedVectors)

			// Segments count 
			fmt.Fprintf(w, "qdrant_collection_segments{collection=\"%s\"} %d\n",
				collectionName, info.Result.SegmentsCount)

			// Status 
			fmt.Fprintf(w, "qdrant_collection_status{collection=\"%s\"} %d\n", 
				collectionName, statusToNumber(info.Result.Status))
		}
	})

	fmt.Println("server starting on http://localhost:9999")
	fmt.Println("Qdrant should be running on http://localhost:6333")
	fmt.Println("\n--- PRESS Ctrl+C to stop ---\n")

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
