package collector

import (
	"log"
	"sync"

	"github.com/MaxFedotov/orcus-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

// XtradbCollector collects Xtradb cluster metrics. It implements prometheus.Collector interface.
type XtradbCollector struct {
	xtradbClient *client.XtradbClient
	metrics      map[string]*prometheus.Desc
	upMetric     prometheus.Gauge
	mutex        sync.Mutex
}

// NewXtradbCollector creates an XtradbCollector.
func NewXtradbCollector(xtradbClient *client.XtradbClient, namespace string) *XtradbCollector {
	return &XtradbCollector{
		xtradbClient: xtradbClient,
		metrics: map[string]*prometheus.Desc{
			"cluter_size":    newGlobalMetric(namespace, "cluter_size", "Number of nodes in Xtradb cluster"),
			"node_state":     newGlobalMetric(namespace, "node_state", "State code of Xtradb cluster node"),
			"cluster_status": newGlobalMetric(namespace, "cluster_status", "State code of Xtradb cluster status"),
		},
		upMetric: newUpMetric(namespace),
	}
}

// Describe sends the super-set of all possible descriptors of Xtradb cluster metrics
// to the provided channel.
func (c *XtradbCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()

	for _, m := range c.metrics {
		ch <- m
	}
}

// Collect fetches metrics from Xtradb cluster and sends them to the provided channel.
func (c *XtradbCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects
	defer c.mutex.Unlock()

	stats, err := c.xtradbClient.GetMetrics()
	if err != nil {
		c.upMetric.Set(serviceDown)
		ch <- c.upMetric
		log.Printf("Error getting Xtradb cluster stats: %v", err)
		return
	}

	c.upMetric.Set(serviceUp)
	ch <- c.upMetric

	ch <- prometheus.MustNewConstMetric(c.metrics["cluter_size"],
		prometheus.GaugeValue, float64(stats.ClusterSize))
	ch <- prometheus.MustNewConstMetric(c.metrics["node_state"],
		prometheus.GaugeValue, float64(stats.NodeState))
	ch <- prometheus.MustNewConstMetric(c.metrics["cluster_status"],
		prometheus.GaugeValue, float64(stats.ClusterStatus))
}
