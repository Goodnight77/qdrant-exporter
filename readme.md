# qdrant-exporter

a Prometheus exporter for Qdrant vector database that exposes per-collection metrics.

## what it does

- scrapes Qdrant API for collection data
- exposes metrics at `/metrics` endpoint
- compatible with Prometheus and Grafana

## how to run

### step 1: start the exporter

```bash
cd simple
go run .
```

the exporter will start on `http://localhost:9999`

### step 2: verify metrics

visit `http://localhost:9999/metrics` to see all metrics

you should see:
```
# HELP qdrant_up Whether Qdrant is reachable
# TYPE qdrant_up gauge
qdrant_up 1

# HELP qdrant_collection_points number of points in the collection
# TYPE qdrant_collection_points gauge
qdrant_collection_points{collection="my-collection"} 100

# HELP qdrant_collection_vectors number of vectors in the collection
# TYPE qdrant_collection_vectors gauge
qdrant_collection_vectors{collection="my-collection"} 200

# HELP qdrant_collection_indexed_vectors number of indexed vectors in the collection
# TYPE qdrant_collection_indexed_vectors gauge
qdrant_collection_indexed_vectors{collection="my-collection"} 200

# HELP qdrant_collection_segments number of segments in the collection
# TYPE qdrant_collection_segments gauge
qdrant_collection_segments{collection="my-collection"} 6

# HELP qdrant_collection_status status of the collection (1=green, 2=yellow, 3=red)
# TYPE qdrant_collection_status gauge
qdrant_collection_status{collection="my-collection"} 1
```

## metrics

| metric name | type | description |
|-------------|------|-------------|
| `qdrant_up` | gauge | whether Qdrant is reachable (1=up, 0=down) |
| `qdrant_collection_points` | gauge | number of points in collection |
| `qdrant_collection_vectors` | gauge | number of vectors in collection |
| `qdrant_collection_indexed_vectors` | gauge | number of indexed vectors |
| `qdrant_collection_segments` | gauge | number of segments |
| `qdrant_collection_status` | gauge | collection status (1=green, 2=yellow, 3=red) |



## how it works

```python
client := NewQdrantClient(at localhost: 6333)
collector := NewQdrantCollector(client)
prometheus.MustRegister(collector) // resgister collector 

http.Handle("/metrics", promhttp.Handler())

```

```
┌──────────────┐
│  Qdrant     │  vector database at :6333
└──────┬───────┘
       │
       │ exporter scrapes
       ↓
┌──────────────┐
│  exporter    │  go app at :9999
│  /metrics    │  exposes prometheus metrics
└──────┬───────┘
       │
       │ prometheus scrapes every 15s
       ↓
┌──────────────┐
│  Prometheus  │  at :9090
│  /graph      │  query and visualize metrics
└──────┬───────┘
       │
       │ grafana queries prometheus
       ↓
┌──────────────┐
│  Grafana     │  dashboards and graphs
└──────────────┘
```



## endpoints
```
| url | what it is | shows |
|------|-------------|--------|
| `http://localhost:9999/metrics` | my exporter | exporter's metrics (qdrant_*) |
| `http://localhost:9090/metrics` | prometheus itself | prometheus internal metrics (memory, cpu, etc.) |
| `http://localhost:9090/graph` | prometheus ui | query and visualize your data |
```

## requirements

- go 1.23+
- qdrant running on `http://localhost:6333`
- prometheus (optional, for scraping metrics)
