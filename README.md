# Qdrant Exporter

A Prometheus exporter for Qdrant vector database that enriches observability with per-collection granularity and Cloud support.

## 🚀 Why is this useful?

Native Qdrant metrics are great but often lack the **per-collection** depth needed for complex applications. This exporter allows you to:
1. **Monitor Multi-tenancy**: Track exactly how many vectors each user collection consumes.
2. **Alert on Shard Health**: Get notified immediately if a shard moves from `active` to `initializing` or `dead`.
3. **Cloud Visibility**: Connect to Qdrant Cloud to monitor managed clusters alongside your local infrastructure.
4. **Capacity Planning**: Use historical vector growth in Prometheus to predict when you'll need to upgrade your Qdrant cluster.

## 🛠️ Key Implementation Changes

- **Qdrant Cloud Support**: Added `api-key` header support to work with GCP/Azure managed clusters.
- **Environment Auto-load**: Integrated `viper` to automatically read from `.env` files.
- **Dual-Mode Operation**: Fixed Windows/WSL/Docker networking to allow scraping local host processes from inside containers.
- **Structured Logging**: Switched to `uber-go/zap` for production-grade, JSON-formatted logs.
- **Robustness**: Added a `/health` endpoint and `scrape_errors` counters to monitor the exporter's own reliability.

## 📦 Quick Start

### 1. Configure Credentials
Create a `.env` file in this directory:
```env
QDRANT_URL=https://your-cluster.qdrant.io:6333
QDRANT_API_KEY=your-secret-key
```

### 2. Run with Docker (Recommended)
This starts the exporter on port **9092** and Prometheus on **9097**:
```bash
docker-compose up --build
```

- **Metrics**: [http://localhost:9092/metrics](http://localhost:9092/metrics)
- **Prometheus UI**: [http://localhost:9097](http://localhost:9097)
---
- **Qdrant Dashboard**: [http://localhost:6333/dashboard](http://localhost:6333/dashboard) (if running Qdrant locally)

### 3. Run Manually (WSL/Linux)
```bash
go run . --listen-address=:9092
```

## 📊 Metrics Documentation

| Metric Name | Type | Labels | Description |
|-------------|------|--------|-------------|
| `qdrant_collection_vectors_total` | Gauge | `collection` | Total vectors. |
| `qdrant_collection_points_total` | Gauge | `collection` | Total points. |
| `qdrant_collection_shards_total` | Gauge | `collection`, `status` | Shards by status (active, dead, etc). |
| `qdrant_exporter_scrape_errors_total` | Counter | `type` | Failures during Qdrant API calls. |
| `qdrant_exporter_last_scrape_success` | Gauge | - | Unix timestamp of last successful check. |

## 🐍 Usage in Python Projects

While this is written in Go, its data is meant for consumption by any language. Here is how you can use it in Python:

### A. Simple Monitoring (via Prometheus API)
If your Python app needs to know its own storage usage to enforce quotas:
```python
import requests

def get_vector_count(collection_name):
    # Query the Prometheus instance we just set up
    resp = requests.get("http://localhost:9097/api/v1/query", 
                        params={'query': f'qdrant_collection_vectors_total{{collection="{collection_name}"}}'})
    data = resp.json()
    return data['data']['result'][0]['value'][1]

print(f"Current usage: {get_vector_count('documents')} vectors")
```

### B. Self-Healing logic
Use the `qdrant_collection_shards_total` metric in your Python management scripts to trigger a collection "repair" or "optimize" if shards go dead.

## 🗺️ Roadmap: Suggested Metrics

To further enhance observability, we plan to implement the following metrics in future versions:

### 1. Performance & Latency
- **`qdrant_collection_query_latency_seconds`**: Histogram of search response times (p50, p95, p99).
- **`qdrant_collection_rps_total`**: Requests per second (RPS) for search, upsert, and delete operations.

### 2. Resource Deep-Dive
- **`qdrant_collection_disk_usage_bytes`**: Actual disk space consumed by segments and payloads.
- **`qdrant_collection_memory_resident_ratio`**: Percentage of the collection currently resident in RAM vs. on-disk.
- **`qdrant_collection_payload_schema_total`**: Number of indexed fields in the payload.

### 3. Maintenance & Health
- **`qdrant_collection_optimizing`**: Boolean (0/1) indicating if a background segment merge is in progress.
- **`qdrant_collection_wal_size_bytes`**: Size of the Write Ahead Log (pending commits).
- **`qdrant_collection_snapshots_total`**: Total number of snapshots/backups available for the collection.

### 4. Cluster & Versioning
- **`qdrant_cluster_peers_total`**: Number of healthy nodes in the cluster.
- **`qdrant_cluster_consensus_latency_seconds`**: Time taken for cluster-wide write agreement.
- **`qdrant_info`**: Static metric exposing Qdrant version and environment metadata.

## 🔮 Future Enhancements

- [ ] **Grafana Dashboard**: Add a pre-built JSON dashboard for one-click visualization.
- [ ] **Remote Write**: Support pushing metrics directly to Managed Prometheus services (AWS AMP, GCP).
- [ ] **Alerting Rules**: Pre-configure Prometheus alerting rules for "Collection Full" or "High Latency".
- [ ] **Auto-Discovery**: Support for dynamic Qdrant cluster discovery in Kubernetes.
