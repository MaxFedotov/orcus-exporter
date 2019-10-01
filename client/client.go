package client

import (
	"log"
	"time"
)

// CreateClientWithRetries tries to create a client for service and retries in case of error
func CreateClientWithRetries(service string, getClient func() (interface{}, error), retries uint, retryInterval time.Duration) (interface{}, error) {
	var err error
	var client interface{}

	for i := 0; i <= int(retries); i++ {
		client, err = getClient()
		if err == nil {
			return client, nil
		}
		if i < int(retries) {
			log.Printf("Could not create %s Client. Retrying in %v...", service, retryInterval)
			time.Sleep(retryInterval)
		}
	}
	return nil, err
}
