package main

import (
	"encoding/json" //parse json resps from qdrant 
	"fmt" // for err formatting 
	"io"
	"net/http" // http reqs
)

// qdrantClient wraps Qdrant HTTP API calls
type QdrantClient struct {
	baseURL string
	apiKey  string
	client  *http.Client 
}

// newQdrantClient creates a new Qdrant client
func NewQdrantClient(url string, apiKey string) *QdrantClient {
	return &QdrantClient{ // mem @
		baseURL: url,
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

// doRequest sends a request to Qdrant
func (qc *QdrantClient) doRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, qc.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if qc.apiKey != "" {
		req.Header.Set("api-key", qc.apiKey)
	}

	resp, err := qc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Qdrant API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("qdrant API returned status: %s body: %s", resp.Status, string(body))
	}

	return resp, nil
}

// GetCollections returns list of collection names
func (qc *QdrantClient) GetCollections() ([]string, error) {
	resp, err := qc.doRequest("/collections")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // close resp body when func exists 

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
	resp, err := qc.doRequest(fmt.Sprintf("/collections/%s", name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var collectionInfoResp CollectionInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionInfoResp); err != nil {
		return nil, fmt.Errorf("failed to decode Qdrant response: %w", err)
	}

	return &collectionInfoResp, nil
}
