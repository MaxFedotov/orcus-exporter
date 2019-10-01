package collector

import (
	"log"
	"sync"

	"github.com/MaxFedotov/orcus-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

// Oauth2ProxyCollector collects oauth2_proxy metrics. It implements prometheus.Collector interface.
type Oauth2ProxyCollector struct {
	oauth2ProxyClient *client.Oauth2ProxyClient
	upMetric          prometheus.Gauge
	mutex             sync.Mutex
}

// NewOauth2ProxyCollector creates an Oauth2ProxyCollector.
func NewOauth2ProxyCollector(oauth2ProxyClient *client.Oauth2ProxyClient, namespace string) *Oauth2ProxyCollector {
	return &Oauth2ProxyCollector{
		oauth2ProxyClient: oauth2ProxyClient,
		upMetric:          newUpMetric(namespace),
	}
}

// Describe sends the super-set of all possible descriptors of oauth2_proxy metrics
// to the provided channel.
func (c *Oauth2ProxyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric.Desc()
}

// Collect fetches metrics from oauth2_proxy and sends them to the provided channel.
func (c *Oauth2ProxyCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.oauth2ProxyClient.GetStatus()
	if err != nil {
		c.upMetric.Set(serviceDown)
		ch <- c.upMetric
		log.Printf("Error getting oauth2_proxy stats: %v", err)
		return
	}

	c.upMetric.Set(serviceUp)
	ch <- c.upMetric
}
