package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	applicationError "github.com/Ecube-Labs/kafka-connect-exporter/pkg/application-error"
)

type Collector struct {
	client *http.Client
}

func New(client *http.Client) *Collector {
	return &Collector{
		client: client,
	}
}

// Get a list of kafka connect connectors from the given host
func (c *Collector) GetConnectors(host string) ([]string, error) {
	response, err := c.client.Get(fmt.Sprintf("%s/connectors", host))
	if err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
		}
		return nil, applicationError.New(response.StatusCode, fmt.Sprintf("Failed to get connectors. status: %d, body: %s", response.StatusCode, string(body)), "")
	}

	var connectors []string
	if err := json.NewDecoder(response.Body).Decode(&connectors); err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}

	return connectors, nil
}

// Retrieve the status of a kafka connect connector
func (c *Collector) GetConnectorStatus(host string, connector string) (*connectorStatus, error) {
	encodedConnectorName := url.PathEscape(connector)
	response, err := c.client.Get(fmt.Sprintf("%s/connectors/%s/status", host, encodedConnectorName))
	if err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
		}
		return nil, applicationError.New(response.StatusCode, fmt.Sprintf("Failed to get connector status. status: %d, body: %s", response.StatusCode, string(body)), "")
	}

	var status connectorStatus
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}

	return &status, nil
}
