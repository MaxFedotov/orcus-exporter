package collector

import "github.com/prometheus/client_golang/prometheus"

const serviceUp = 1
const serviceDown = 0

func newGlobalMetric(namespace string, metricName string, docString string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, docString, nil, nil)
}

func newUpMetric(namespace string) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "up",
		Help:      "Status of the last metric scrape",
	})
}
