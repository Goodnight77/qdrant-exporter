package main

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type QdrantCollector struct {
	client            *QdrantClient
	logger            *zap.Logger
	vectorsTotalDesc  *prometheus.Desc
	pointsTotalDesc   *prometheus.Desc
	shardsTotalDesc   *prometheus.Desc
	indexedVectorsDesc *prometheus.Desc
	
	scrapeErrors      *prometheus.CounterVec
	lastScrapeSuccess prometheus.Gauge
	
	mu                sync.RWMutex
	lastSuccessTime   time.Time
}

func NewQdrantCollector(client *QdrantClient, logger *zap.Logger) *QdrantCollector {
	return &QdrantCollector{
		client: client,
		logger: logger,
		vectorsTotalDesc: prometheus.NewDesc(
			"qdrant_collection_vectors_total",
			"Total number of vectors in the collection",
			[]string{"collection"},
			nil,
		),
		pointsTotalDesc: prometheus.NewDesc(
			"qdrant_collection_points_total",
			"Total number of points in the collection",
			[]string{"collection"},
			nil,
		),
		shardsTotalDesc: prometheus.NewDesc(
			"qdrant_collection_shards_total",
			"Count of shards by status",
			[]string{"collection", "status"},
			nil,
		),
		indexedVectorsDesc: prometheus.NewDesc(
			"qdrant_collection_indexed_vectors",
			"Count of indexed vectors vs total vectors",
			[]string{"collection", "vector_name"},
			nil,
		),
		scrapeErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "qdrant_exporter_scrape_errors_total",
				Help: "Total number of scrape errors",
			},
			[]string{"type"},
		),
		lastScrapeSuccess: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "qdrant_exporter_last_scrape_success",
				Help: "Timestamp of the last successful scrape",
			},
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (qc *QdrantCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- qc.vectorsTotalDesc
	ch <- qc.pointsTotalDesc
	ch <- qc.shardsTotalDesc
	ch <- qc.indexedVectorsDesc
	qc.scrapeErrors.Describe(ch)
	qc.lastScrapeSuccess.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (qc *QdrantCollector) Collect(ch chan<- prometheus.Metric) {
	collections, err := qc.client.GetCollections()
	if err != nil {
		qc.logger.Error("Error fetching collections for metrics", zap.Error(err))
		qc.scrapeErrors.WithLabelValues("collections").Inc()
		qc.scrapeErrors.Collect(ch)
		qc.lastScrapeSuccess.Collect(ch)
		return
	}

	success := true
	for _, name := range collections {
		// Collection Info
		info, err := qc.client.GetCollectionInfo(name)
		if err != nil {
			qc.logger.Error("Error fetching collection info", zap.String("collection", name), zap.Error(err))
			qc.scrapeErrors.WithLabelValues("collection_info").Inc()
			success = false
		} else {
			ch <- prometheus.MustNewConstMetric(
				qc.vectorsTotalDesc,
				prometheus.GaugeValue,
				float64(info.VectorsCount),
				name,
			)
			ch <- prometheus.MustNewConstMetric(
				qc.pointsTotalDesc,
				prometheus.GaugeValue,
				float64(info.PointsCount),
				name,
			)
			ch <- prometheus.MustNewConstMetric(
				qc.indexedVectorsDesc,
				prometheus.GaugeValue,
				float64(info.VectorsCount),
				name,
				"default",
			)
		}

		// Cluster Info (Shards)
		clusterInfo, err := qc.client.GetCollectionClusterInfo(name)
		if err != nil {
			qc.logger.Debug("Collection not in cluster mode or failed", zap.String("collection", name), zap.Error(err))
		} else {
			statusCounts := make(map[string]int)
			for _, s := range clusterInfo.Shards {
				statusCounts[s.Status]++
			}
			for status, count := range statusCounts {
				ch <- prometheus.MustNewConstMetric(
					qc.shardsTotalDesc,
					prometheus.GaugeValue,
					float64(count),
					name,
					status,
				)
			}
		}
	}

	if success {
		now := time.Now()
		qc.mu.Lock()
		qc.lastSuccessTime = now
		qc.mu.Unlock()
		qc.lastScrapeSuccess.Set(float64(now.Unix()))
	}

	qc.scrapeErrors.Collect(ch)
	qc.lastScrapeSuccess.Collect(ch)
}

func (qc *QdrantCollector) GetLastSuccessTime() time.Time {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return qc.lastSuccessTime
}
