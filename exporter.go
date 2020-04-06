package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/MaxFedotov/orcus-exporter/client"
	"github.com/MaxFedotov/orcus-exporter/collector"
	nginxclient "github.com/nginxinc/nginx-prometheus-exporter/client"
	nginxcollector "github.com/nginxinc/nginx-prometheus-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version            = "dev"
	commit             = "none"
	date               = "unknown"
	listenAddress      = flag.String("web.listen-address", ":9114", "Address to listen on for web interface")
	metricsPath        = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics")
	retries            = flag.Uint("config.retries", 0, "Number of retries the exporter will make on start in order to inialize collectors")
	retryInterval      = flag.Duration("config.retry-interval", time.Second*5, "Interval between retries to connect to collectors endpoint")
	timeout            = flag.Duration("config.timeout", time.Second*5, "Timeout for scraping metrics for collector")
	sslVerify          = flag.Bool("config.ssl-verify", false, "Verify SSL certificates")
	nginx              = flag.Bool("collector.nginx", true, "Collect data for nginx")
	nginxURI           = flag.String("collector.nginx.uri", "http://127.0.0.1:80/nginx_status", "URI for scraping nginx metrics")
	oauth2Proxy        = flag.Bool("collector.oauth2_proxy", true, "Collect data for oauth2_proxy")
	oauth2ProxyURI     = flag.String("collector.oauth2_proxy.uri", "http://127.0.0.1:4180/ping", "URI for scraping oauth2_proxy metrics")
	orcus              = flag.Bool("collector.orcus", true, "Collect data for orcus")
	orcusURI           = flag.String("collector.orcus.uri", "http://127.0.0.1:3008/metrics", "URI for scraping orcus metrics")
	orchestrator       = flag.Bool("collector.orchestrator", true, "Collect data for orchestrator")
	orchestratorURI    = flag.String("collector.orchestrator.uri", "http://127.0.0.1:3000/api", "URI for scraping orchestrator metrics")
	xtradbCluster      = flag.Bool("collector.xtradb-cluster", true, "Collect data for XtraDB cluster")
	xtradbClusterMycnf = flag.String("collector.xtradb-cluster.my-cnf", path.Join(os.Getenv("HOME"), ".my.cnf"), "Path to .my.cnf file to read MySQL credentials from")
)

func main() {
	flag.Parse()
	log.Printf("Starting Orcus Prometheus Exporter Version=%v GitCommit=%v Date=%v", version, commit, date)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	go func() {
		log.Printf("SIGTERM received: %v. Exiting...", <-signalChan)
		os.Exit(0)
	}()

	registry := prometheus.NewRegistry()

	buildInfoMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "orcusexporter_build_info",
			Help: "Exporter build information",
			ConstLabels: prometheus.Labels{
				"version":   version,
				"gitCommit": commit,
				"date":      date,
			},
		},
	)
	buildInfoMetric.Set(1)

	registry.MustRegister(buildInfoMetric)

	httpClient := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !*sslVerify},
		},
	}

	if *nginx {
		service := "nginx"
		nginxClient, err := client.CreateClientWithRetries(service, func() (interface{}, error) {
			return nginxclient.NewNginxClient(httpClient, *nginxURI)
		}, *retries, *retryInterval)
		if err != nil {
			log.Fatalf("Could not create Nginx Client: %v", err)
		}
		registry.MustRegister(nginxcollector.NewNginxCollector(nginxClient.(*nginxclient.NginxClient), service))
	}

	if *oauth2Proxy {
		service := "oauth2_proxy"
		oauth2ProxyClient, err := client.CreateClientWithRetries(service, func() (interface{}, error) {
			return client.NewOauth2ProxyClient(httpClient, *oauth2ProxyURI)
		}, *retries, *retryInterval)
		if err != nil {
			log.Fatalf("Could not create oauth2_proxy Client: %v", err)
		}
		registry.MustRegister(collector.NewOauth2ProxyCollector(oauth2ProxyClient.(*client.Oauth2ProxyClient), service))
	}

	if *orcus {
		service := "orcus"
		orcusClient, err := client.CreateClientWithRetries(service, func() (interface{}, error) {
			return client.NewOrcusClient(httpClient, *orcusURI)
		}, *retries, *retryInterval)
		if err != nil {
			log.Fatalf("Could not create Orcus Client: %v", err)
		}
		registry.MustRegister(collector.NewOrcusCollector(orcusClient.(*client.OrcusClient), service))
	}

	if *orchestrator {
		service := "orchestrator"
		orchestatorClient, err := client.CreateClientWithRetries(service, func() (interface{}, error) {
			return client.NewOrchestratorClient(httpClient, *orchestratorURI)
		}, *retries, *retryInterval)
		if err != nil {
			log.Fatalf("Could not create Orchestrator Client: %v", err)
		}
		registry.MustRegister(collector.NewOrchestratorCollector(orchestatorClient.(*client.OrchestratorClient), service))
	}

	if *xtradbCluster {
		service := "xtradb_cluster"
		xtradbClient, err := client.CreateClientWithRetries(service, func() (interface{}, error) {
			return client.NewXtradbClient(*xtradbClusterMycnf, *sslVerify)
		}, *retries, *retryInterval)
		if err != nil {
			log.Fatalf("Could not create Xtradb cluster Client: %v", err)
		}
		registry.MustRegister(collector.NewXtradbCollector(xtradbClient.(*client.XtradbClient), service))
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>Orcus Exporter</title></head>
			<body>
			<h1>Orcus Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Printf("Error while sending a response for the '/' path: %v", err)
		}
	})
	log.Printf("Orcus Prometheus Exporter has successfully started")
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
