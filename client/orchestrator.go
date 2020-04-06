package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OrchestratorClient allows you to get Orchestrator metrics.
type OrchestratorClient struct {
	apiEndpoint string
	httpClient  *http.Client
}

// OrchestratorFailovers represents failover metrics
type OrchestratorFailovers struct {
	ID int64
}

// OrchestratorFailedSeed represents failed seeds
type OrchestratorFailedSeed struct {
	SeedID int64
}

// OrchestratorMetrics represents Orchestrator metrics.
type OrchestratorMetrics struct {
	Status         HealthStatus
	Problems       []interface{}
	LastFailoverID int64
	FailedSeeds    int
}

// HealthStatus represents status related metrics.
type HealthStatus struct {
	Details struct {
		Healthy        bool
		IsActiveNode   bool
		AvailableNodes []interface{}
	}
}

// NewOrchestratorClient creates an OrchestratorClient.
func NewOrchestratorClient(httpClient *http.Client, apiEndpoint string) (*OrchestratorClient, error) {
	client := &OrchestratorClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
	}

	if _, err := client.GetMetrics(); err != nil {
		return nil, fmt.Errorf("Failed to create Orchestrator client: %v", err)
	}

	return client, nil
}

// GetMetrics fetches Orchestrator metrics.
func (client *OrchestratorClient) GetMetrics() (*OrchestratorMetrics, error) {
	var metrics OrchestratorMetrics
	var err error
	metrics.Status, err = client.getStatus()
	if err != nil {
		return nil, err
	}
	metrics.Problems, err = client.getProblems("/problems")
	if err != nil {
		return nil, err
	}
	metrics.LastFailoverID, err = client.getFailovers("/audit-failure-detection")
	if err != nil {
		return nil, err
	}
	metrics.FailedSeeds, err = client.getFailedSeeds("/agents-failed-seeds")
	if err != nil {
		return nil, err
	}
	return &metrics, nil

}

func (client *OrchestratorClient) getFailedSeeds(endpoint string) (int, error) {
	url := client.apiEndpoint + endpoint
	resp, err := client.httpClient.Get(url)
	failedSeeds := []OrchestratorFailedSeed{}
	if err != nil {
		return 0, fmt.Errorf("failed to get %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read the response body: %v", err)
	}

	err = json.Unmarshal(body, &failedSeeds)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}
	return len(failedSeeds), nil
}

func (client *OrchestratorClient) getFailovers(endpoint string) (lastFailoverID int64, err error) {
	url := client.apiEndpoint + endpoint
	resp, err := client.httpClient.Get(url)
	failovers := []OrchestratorFailovers{}
	if err != nil {
		return 0, fmt.Errorf("failed to get %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read the response body: %v", err)
	}

	err = json.Unmarshal(body, &failovers)
	if err != nil {
		return 0, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}
	for _, failover := range failovers {
		if failover.ID > lastFailoverID {
			lastFailoverID = failover.ID
		}
	}
	return lastFailoverID, nil
}

func (client *OrchestratorClient) getProblems(endpoint string) (metric []interface{}, err error) {
	url := client.apiEndpoint + endpoint
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return metric, fmt.Errorf("failed to get %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return metric, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return metric, fmt.Errorf("failed to read the response body: %v", err)
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		return metric, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}
	return metric, nil
}

func (client *OrchestratorClient) getStatus() (metric HealthStatus, err error) {
	url := client.apiEndpoint + "/status"
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return metric, fmt.Errorf("failed to get %v: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return metric, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return metric, fmt.Errorf("failed to read the response body: %v", err)
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		return metric, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}
	return metric, nil
}
