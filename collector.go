package main

import (
	// "fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// qdrantCollector collects metrics from Qdrant and exposes them to Prometheus
type qdrantCollector struct {
	client *QdrantClient // point to qdrantclient to fetch data 

	// metric descriptors
	pointsDesc    *prometheus.Desc // descriptor : what metric exist 
	vectorsDesc   *prometheus.Desc
	indexedDesc   *prometheus.Desc
	segmentsDesc  *prometheus.Desc
	statusDesc    *prometheus.Desc
}

// newQdrantCollector creates a new collector
func NewQdrantCollector(client *QdrantClient) *qdrantCollector { // return a qdradntCollector : poniter 
	return &qdrantCollector{ // return @ of struct 
		client: client,
		pointsDesc: prometheus.NewDesc(
			"qdrant_collection_points", // metric name 
			"number of points in the collection", // HELP text (description)
			[]string{"collection"}, // label 
			nil, // const lables (none here)
		),
		vectorsDesc: prometheus.NewDesc(
			"qdrant_collection_vectors",
			"number of vectors in the collection",
			[]string{"collection"},
			nil,
		),
		indexedDesc: prometheus.NewDesc(
			"qdrant_collection_indexed_vectors",
			"number of indexed vectors in the collection",
			[]string{"collection"},
			nil,
		),
		segmentsDesc: prometheus.NewDesc(
			"qdrant_collection_segments",
			"number of segments in the collection",
			[]string{"collection"},
			nil,
		),
		statusDesc: prometheus.NewDesc(
			"qdrant_collection_status",
			"status of the collection (1=green, 2=yellow, 3=red)",
			[]string{"collection"},
			nil,
		),
	}
}

// describe implements prometheus.Collector interface
// sends metric descriptors to Prometheus
func (qc *qdrantCollector) Describe(ch chan<- *prometheus.Desc) { // ch : channel sends to prometheus
	ch <- qc.pointsDesc
	ch <- qc.vectorsDesc
	ch <- qc.indexedDesc
	ch <- qc.segmentsDesc
	ch <- qc.statusDesc
}

// collect implements prometheus.Collector interface
// scrapes Qdrant and sends metrics to Prometheus
func (qc *qdrantCollector) Collect(ch chan<- prometheus.Metric) { // ch for metric vzlues 
	// get all collections
	collections, err := qc.client.GetCollections()
	if err != nil {
		// no collections, emit no metrics
		return
	}

	// for each collection, get info and emit metrics
	for _, name := range collections {
		info, err := qc.client.GetCollectionInfo(name)
		if err != nil {
			// skip this collection if error
			continue
		}

		// emit metrics
		ch <- prometheus.MustNewConstMetric(
			qc.pointsDesc, // which metric desc
			prometheus.GaugeValue, // type gauge 
			float64(info.Result.PointsCount), // value 
			name, // label 
		)
		ch <- prometheus.MustNewConstMetric(
			qc.vectorsDesc,
			prometheus.GaugeValue,
			float64(info.Result.VectorsCount),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			qc.indexedDesc,
			prometheus.GaugeValue,
			float64(info.Result.IndexedVectors),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			qc.segmentsDesc,
			prometheus.GaugeValue,
			float64(info.Result.SegmentsCount),
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			qc.statusDesc,
			prometheus.GaugeValue,
			float64(statusToNumber(info.Result.Status)),
			name,
		)
	}
}
