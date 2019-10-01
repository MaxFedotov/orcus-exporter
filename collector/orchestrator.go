package collector

import (
	"log"
	"sync"

	"github.com/MaxFedotov/orcus-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

// OrchestratorCollector collects Orchestrator metrics. It implements prometheus.Collector interface.
type OrchestratorCollector struct {
	orchestratorClient *client.OrchestratorClient
	metrics            map[string]*prometheus.Desc
	upMetric           prometheus.Gauge
	mutex              sync.Mutex
}

// NewOrchestratorCollector creates an OrchestratorCollector.
func NewOrchestratorCollector(orchestratorClient *client.OrchestratorClient, namespace string) *OrchestratorCollector {
	return &OrchestratorCollector{
		orchestratorClient: orchestratorClient,
		metrics: map[string]*prometheus.Desc{
			"cluter_size":      newGlobalMetric(namespace, "cluter_size", "Number of nodes in Orchestrator cluster"),
			"is_active_node":   newGlobalMetric(namespace, "is_active_node", "If this node is active Orchestrator node"),
			"problems":         newGlobalMetric(namespace, "problems", "Count of MySQL clusters with problems"),
			"last_failover_id": newGlobalMetric(namespace, "last_failover_id", "ID of last failover"),
			"is_healthy":       newGlobalMetric(namespace, "is_healthy", "Orchestrator node health status"),
		},
		upMetric: newUpMetric(namespace),
	}
}

// Describe sends the super-set of all possible descriptors of Orchestrator metrics
// to the provided channel.
func (c *OrchestratorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()

	for _, m := range c.metrics {
		ch <- m
	}
}

func boolToFloat64(val bool) float64 {
	if val {
		return 1.0
	}
	return 0
}

// Collect fetches metrics from Orchestrator and sends them to the provided channel.
func (c *OrchestratorCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects
	defer c.mutex.Unlock()

	stats, err := c.orchestratorClient.GetMetrics()
	if err != nil {
		c.upMetric.Set(serviceDown)
		ch <- c.upMetric
		log.Printf("Error getting Orchestrator stats: %v", err)
		return
	}

	c.upMetric.Set(serviceUp)
	ch <- c.upMetric

	ch <- prometheus.MustNewConstMetric(c.metrics["cluter_size"],
		prometheus.GaugeValue, float64(len(stats.Status.Details.AvailableNodes)))
	ch <- prometheus.MustNewConstMetric(c.metrics["is_active_node"],
		prometheus.GaugeValue, boolToFloat64(stats.Status.Details.IsActiveNode))
	ch <- prometheus.MustNewConstMetric(c.metrics["problems"],
		prometheus.GaugeValue, float64(len(stats.Problems)))
	ch <- prometheus.MustNewConstMetric(c.metrics["last_failover_id"],
		prometheus.CounterValue, float64(stats.LastFailoverID))
	ch <- prometheus.MustNewConstMetric(c.metrics["is_healthy"],
		prometheus.GaugeValue, boolToFloat64(stats.Status.Details.Healthy))
}
