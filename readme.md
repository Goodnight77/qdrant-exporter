# qdrant-exporter

A Prometheus exporter for Qdrant vector database that enriches the native /metrics endpoint with per-collection metrics and custom observability data.

## main flow 

```python
client := NewQdrantClient(at localhost: 6333)
collector := NewQdrantCollector(client)
prometheus.MustRegister(collector) // resgister collector 

http.Handle("/metrics", promhttp.Handler())

```
```markdown
 Prometheus scrapes /metrics             
     ↓                                 
   promhttp.Handler()                    
     ↓                                  
    Calls collector.Collect()          
     ↓                                  
   collector gets data from client          
     ↓                                  
   Metrics sent to /metrics response 
   ```