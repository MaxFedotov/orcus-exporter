package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OrcusClient allows you to get Orcus metrics.
type OrcusClient struct {
	apiEndpoint string
	httpClient  *http.Client
}

// OrcusMetrics represents Orcus metrics.
type OrcusMetrics struct {
	LastSyncDurationSeconds float64
	TotalSyncClusters       uint64
	TotalSyncErrors         uint64
	TotalSyncCount          uint64
}

// NewOrcusClient creates an OrcusClient.
func NewOrcusClient(httpClient *http.Client, apiEndpoint string) (*OrcusClient, error) {
	client := &OrcusClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
	}

	if _, err := client.GetMetrics(); err != nil {
		return nil, fmt.Errorf("Failed to create Orcus client: %v", err)
	}

	return client, nil
}

// GetMetrics fetches Orcus metrics.
func (client *OrcusClient) GetMetrics() (*OrcusMetrics, error) {
	resp, err := client.httpClient.Get(client.apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %v", client.apiEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %v", err)
	}

	var metrics OrcusMetrics
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}

	return &metrics, nil
}
