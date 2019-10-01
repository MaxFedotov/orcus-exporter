package collector

import (
	"log"
	"sync"

	"github.com/MaxFedotov/orcus-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

// OrcusCollector collects Orcus metrics. It implements prometheus.Collector interface.
type OrcusCollector struct {
	orcusClient *client.OrcusClient
	metrics     map[string]*prometheus.Desc
	upMetric    prometheus.Gauge
	mutex       sync.Mutex
}

// NewOrcusCollector creates an OrcusCollector.
func NewOrcusCollector(orcusClient *client.OrcusClient, namespace string) *OrcusCollector {
	return &OrcusCollector{
		orcusClient: orcusClient,
		metrics: map[string]*prometheus.Desc{
			"clusters_synced_total":      newGlobalMetric(namespace, "clusters_synced_total", "Total synced clusters"),
			"sync_errors_total":          newGlobalMetric(namespace, "sync_errors_total", "Total errors during sync"),
			"last_sync_duration_seconds": newGlobalMetric(namespace, "last_sync_duration_seconds", "Duration of last sync process"),
			"sync_count_total":           newGlobalMetric(namespace, "sync_count_total", "Total count of sync tasks"),
		},
		upMetric: newUpMetric(namespace),
	}
}

// Describe sends the super-set of all possible descriptors of Orcus metrics
// to the provided channel.
func (c *OrcusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()

	for _, m := range c.metrics {
		ch <- m
	}
}

// Collect fetches metrics from Orcus and sends them to the provided channel.
func (c *OrcusCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects
	defer c.mutex.Unlock()

	stats, err := c.orcusClient.GetMetrics()
	if err != nil {
		c.upMetric.Set(serviceDown)
		ch <- c.upMetric
		log.Printf("Error getting Orcus stats: %v", err)
		return
	}

	c.upMetric.Set(serviceUp)
	ch <- c.upMetric

	ch <- prometheus.MustNewConstMetric(c.metrics["clusters_synced_total"],
		prometheus.CounterValue, float64(stats.TotalSyncClusters))
	ch <- prometheus.MustNewConstMetric(c.metrics["sync_errors_total"],
		prometheus.CounterValue, float64(stats.TotalSyncErrors))
	ch <- prometheus.MustNewConstMetric(c.metrics["last_sync_duration_seconds"],
		prometheus.GaugeValue, float64(stats.LastSyncDurationSeconds))
	ch <- prometheus.MustNewConstMetric(c.metrics["sync_count_total"],
		prometheus.CounterValue, float64(stats.TotalSyncCount))
}
