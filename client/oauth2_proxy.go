package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Oauth2ProxyClient allows you to get oauth2_proxy metrics.
type Oauth2ProxyClient struct {
	apiEndpoint string
	httpClient  *http.Client
}

// NewOauth2ProxyClient creates an Oauth2ProxyClient.
func NewOauth2ProxyClient(httpClient *http.Client, apiEndpoint string) (*Oauth2ProxyClient, error) {
	client := &Oauth2ProxyClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
	}

	if err := client.GetStatus(); err != nil {
		return nil, fmt.Errorf("Failed to create oauth2_proxy: %v", err)
	}

	return client, nil
}

// GetStatus fetches the oauth2_proxy metrics.
func (client *Oauth2ProxyClient) GetStatus() error {
	resp, err := client.httpClient.Get(client.apiEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get %v: %v", client.apiEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read the response body: %v", err)
	}

	return nil
}
