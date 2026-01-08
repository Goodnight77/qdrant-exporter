package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CollectionInfoResponse struct {
	Result struct {
		VectorsCount int64 `json:"vectors_count"`
		PointsCount  int64 `json:"points_count"`
		Status       string `json:"status"`
		PayloadSchema map[string]interface{} `json:"payload_schema"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

type ClusterInfoResponse struct {
	Result struct {
		PeerID uint64 `json:"peer_id"`
		Shards []struct {
			ShardID uint32 `json:"shard_id"`
			Status  string `json:"status"`
		} `json:"shards"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

type CollectionsResponse struct {
	Result struct {
		Collections []struct {
			Name string `json:"name"`
		} `json:"collections"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

type QdrantClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewQdrantClient(baseURL, apiKey string) *QdrantClient {
	return &QdrantClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (qc *QdrantClient) doRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", qc.baseURL, path), nil)
	if err != nil {
		return nil, err
	}

	if qc.apiKey != "" {
		req.Header.Set("api-key", qc.apiKey)
	}

	return qc.client.Do(req)
}

func (qc *QdrantClient) GetCollections() ([]string, error) {
	resp, err := qc.doRequest("/collections")
	if err != nil {
		return nil, fmt.Errorf("failed to call Qdrant API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant API returned status: %s", resp.Status)
	}

	var collectionsResp CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode Qdrant response: %w", err)
	}

	var collections []string
	for _, col := range collectionsResp.Result.Collections {
		collections = append(collections, col.Name)
	}

	return collections, nil
}

type CollectionInfo struct {
	VectorsCount int64
	PointsCount  int64
	Status       string
}

type ShardInfo struct {
	ShardID uint32
	Status  string
}

type ClusterInfo struct {
	Shards []ShardInfo
}

func (qc *QdrantClient) GetCollectionInfo(name string) (*CollectionInfo, error) {
	resp, err := qc.doRequest(fmt.Sprintf("/collections/%s", name))
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

	return &CollectionInfo{
		VectorsCount: collectionInfoResp.Result.VectorsCount,
		PointsCount:  collectionInfoResp.Result.PointsCount,
		Status:       collectionInfoResp.Result.Status,
	}, nil
}

func (qc *QdrantClient) GetCollectionClusterInfo(name string) (*ClusterInfo, error) {
	resp, err := qc.doRequest(fmt.Sprintf("/collections/%s/cluster", name))
	if err != nil {
		return nil, fmt.Errorf("failed to call Qdrant API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If 404 or other error, might mean not in cluster mode
		return nil, fmt.Errorf("qdrant API returned status: %s", resp.Status)
	}

	var clusterInfoResp ClusterInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&clusterInfoResp); err != nil {
		return nil, fmt.Errorf("failed to decode Qdrant response: %w", err)
	}

	var shards []ShardInfo
	for _, s := range clusterInfoResp.Result.Shards {
		shards = append(shards, ShardInfo{
			ShardID: s.ShardID,
			Status:  s.Status,
		})
	}

	return &ClusterInfo{
		Shards: shards,
	}, nil
}

// Update CollectionInfoResponse to include config for vector size
type CollectionConfigResponse struct {
	Result struct {
		Config struct {
			Params struct {
				Vectors interface{} `json:"vectors"` // Can be map or just size info, handling simple case
			} `json:"params"`
		} `json:"config"`
	} `json:"result"`
}

// ScrollPointsResponse structure for browsing data
type ScrollPointsResponse struct {
	Result struct {
		Points []struct {
			ID      interface{}            `json:"id"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"points"`
	} `json:"result"`
}

type PointData struct {
	ID      string
	Payload string
}

func (qc *QdrantClient) GetCollectionVectorSize(name string) (int, error) {
	resp, err := qc.doRequest(fmt.Sprintf("/collections/%s", name))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var configResp CollectionConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return 0, err
	}
	
	// Qdrant vectors config can be complex (multiple named vectors). 
	// This is a simplified check for single vector size.
	// If it's a map, we'd need more complex parsing.
	// For now, attempting to parse as map or struct is hard without creating specific types.
	// Let's assume standard single vector config for this user request or return 0 if complex.
	
	// A robust implementation would check type of `Vectors`. 
	// For simplicity in this helper:
	if vMap, ok := configResp.Result.Config.Params.Vectors.(map[string]interface{}); ok {
		if size, ok := vMap["size"].(float64); ok {
			return int(size), nil
		}
	}
	return 0, nil
}

func (qc *QdrantClient) ScrollPoints(name string) ([]PointData, error) {
	// We need POST request for scroll, doRequest is GET.
	// Let's make a custom request here.
	// Using GET /points with limit? No, Qdrant uses POST /collections/{name}/points/scroll
	reqBody := `{"limit": 10, "with_payload": true}`
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/collections/%s/points/scroll", qc.baseURL, name), strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if qc.apiKey != "" {
		req.Header.Set("api-key", qc.apiKey)
	}

	resp, err := qc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	var scrollResp ScrollPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&scrollResp); err != nil {
		return nil, err
	}

	var points []PointData
	for _, p := range scrollResp.Result.Points {
		payloadBytes, _ := json.Marshal(p.Payload)
		points = append(points, PointData{
			ID:      fmt.Sprintf("%v", p.ID),
			Payload: string(payloadBytes),
		})
	}
	return points, nil
}
