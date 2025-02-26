package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ecube-labs/kafka-connect-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Should return a new collector", func(t *testing.T) {
		collector := New()
		assert.NotNil(t, collector)
	})
	t.Run("Should return a collector implementing the prometheus.Collector interface", func(t *testing.T) {
		collector := New()
		_, ok := interface{}(collector).(prometheus.Collector)
		assert.True(t, ok)
	})
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestGetConnectors(t *testing.T) {
	t.Run("Should return a list of connectors", func(t *testing.T) {
		roundTripper := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				response.Write([]byte(`["connector1", "connector2"]`))
				return response.Result(), nil
			},
		}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		connectors, err := collector.getConnectors("test")
		assert.Nil(t, err)
		assert.Equal(t, []string{"connector1", "connector2"}, connectors)
	})

	t.Run("Should return an error when the request fails", func(t *testing.T) {
		roundTripper := &mockRoundTripper{
			roundTripFunc: (func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				response.WriteHeader(444)
				response.Write([]byte(`"Internal Server Error"`))
				return response.Result(), nil
			})}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		connectors, err := collector.getConnectors("test")
		assert.Nil(t, connectors)
		assert.NotNil(t, err)
		assert.Equal(t, `Failed to get connectors. status: 444, body: "Internal Server Error"`, err.Error())
	})
}

func TestGetConnectorStatus(t *testing.T) {
	t.Run("Should return the status of a connector", func(t *testing.T) {
		roundTripper := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				mockResponse := connectorStatus{
					Name: "connector1",
					Connector: struct {
						State    string `json:"state"`
						WorkerID string `json:"worker_id"`
					}{State: "RUNNING", WorkerID: "123"},
					Tasks: []connectorTaskStatus{
						{ID: 1, State: "RUNNING", WorkerID: "123", Trace: "trace1"},
					},
				}

				jsonResponse, err := json.Marshal(mockResponse)
				if err != nil {
					panic(err)
				}

				response.Write(jsonResponse)

				return response.Result(), nil
			},
		}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		status, err := collector.getConnectorStatus("test", "connector1")
		assert.Nil(t, err)
		assert.Equal(t, &connectorStatus{Name: "connector1", Connector: struct {
			State    string `json:"state"`
			WorkerID string `json:"worker_id"`
		}{State: "RUNNING", WorkerID: "123"},
			Tasks: []connectorTaskStatus{{ID: 1, State: "RUNNING", WorkerID: "123", Trace: "trace1"}}}, status) // Compare this snippet from internal/collector/types.go:
	})
	t.Run("Should return an error when the request fails", func(t *testing.T) {
		roundTripper := &mockRoundTripper{
			roundTripFunc: (func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				response.WriteHeader(444)
				response.Write([]byte(`"Internal Server Error"`))
				return response.Result(), nil
			})}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		connectors, err := collector.getConnectorStatus("test", "connector1")
		assert.Nil(t, connectors)
		assert.NotNil(t, err)
		assert.Equal(t, `Failed to get connector status. status: 444, body: "Internal Server Error"`, err.Error())
	})
}

func TestCollect(t *testing.T) {
	t.Run("Should collect metrics successfully", func(t *testing.T) {
		mockHosts := []string{"http://test-host1", "http://test-host2"}
		config.KafkaConnectHosts = mockHosts

		mockConnectors := []string{"connector1", "connector2"}
		roundTripper := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()

				if req.URL.Path == "/connectors" {
					resp, _ := json.Marshal(mockConnectors)
					response.Write([]byte(resp))
				} else if req.URL.Path == "/connectors/connector1/status" {
					response.Write([]byte(`{
						"name": "connector1",
						"tasks": [{"state": "RUNNING"}, {"state": "FAILED"}]
					}`))
				} else if req.URL.Path == "/connectors/connector2/status" {
					response.Write([]byte(`{
						"name": "connector2",
						"tasks": [{"state": "PAUSED"}]
					}`))
				} else {
					response.WriteHeader(http.StatusNotFound)
				}

				return response.Result(), nil
			},
		}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		ch := make(chan prometheus.Metric, len(mockHosts)*len(mockConnectors)*5+len(mockHosts))
		go func() {
			collector.Collect(ch)
			close(ch)
		}()

		var collectedMetrics []prometheus.Metric
		count := 0
		for metric := range ch {
			count++
			collectedMetrics = append(collectedMetrics, metric)
		}

		assert.Equal(t, len(mockHosts)*len(mockConnectors)*5+len(mockHosts), len(collectedMetrics))
	})

	t.Run("Should log errors when API fails", func(t *testing.T) {
		mockHosts := []string{"http://test-host1"}
		config.KafkaConnectHosts = mockHosts

		roundTripper := &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte(`"Internal Server Error"`))
				return response.Result(), nil
			},
		}

		collector := New()
		collector.client = &http.Client{Transport: roundTripper}

		ch := make(chan prometheus.Metric, 10)
		go func() {
			collector.Collect(ch)
			close(ch)
		}()

		collected := false
		for range ch {
			collected = true
		}

		assert.False(t, collected)
	})
}
