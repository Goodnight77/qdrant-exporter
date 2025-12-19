package main

import (
	"encoding/json" //parse json resps from qdrant 
	"fmt" // for err formatting 
	"net/http" // http reqs
)

// QdrantClient wraps Qdrant HTTP API calls
type QdrantClient struct {
	baseURL string
	client  *http.Client 
}

// NewQdrantClient creates a new Qdrant client
func NewQdrantClient(url string) *QdrantClient {
	return &QdrantClient{ // mem @
		baseURL: url,
		client:  &http.Client{},
	}
}

// GetCollections returns list of collection names
func (qc *QdrantClient) GetCollections() ([]string, error) {
	resp, err := qc.client.Get(qc.baseURL + "/collections") // http get req to /collections endpoint
	if err != nil {
		return nil, fmt.Errorf("failed to call Qdrant API: %w", err)
	}
	defer resp.Body.Close() // close resp body when func exists 

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant API returned status: %s", resp.Status)
	}

	var collectionsResp CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil { // parse json resp into struct 
		return nil, fmt.Errorf("failed to decode Qdrant response: %w", err)
	}

	var collections []string
	for _, col := range collectionsResp.Result.Collections {
		collections = append(collections, col.Name)
	}

	return collections, nil
}

// GetCollectionInfo returns details for a specific collection
func (qc *QdrantClient) GetCollectionInfo(name string) (*CollectionInfoResponse, error) {
	resp, err := qc.client.Get(fmt.Sprintf("%s/collections/%s", qc.baseURL, name))
	if err != nil {
		return nil, fmt.Errorf("failed to call Qdrant API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant API returned status: %s", resp.Status)
	}

	var collectionInfoResp CollectionInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionInfoResp); err != nil {
		return nil, fmt.Errorf("failed to decode Qdrant response: %w", err)
	}

	return &collectionInfoResp, nil
}
