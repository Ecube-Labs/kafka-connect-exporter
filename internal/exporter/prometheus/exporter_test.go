package exporter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/collector"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Should return a new exporter", func(t *testing.T) {
		collector := collector.New(&http.Client{Transport: &mockRoundTripper{}})
		exporter := New(collector)
		assert.NotNil(t, exporter)
	})
	t.Run("Should return a exporter implementing the prometheus.Collector interface", func(t *testing.T) {
		collector := collector.New(&http.Client{Transport: &mockRoundTripper{}})
		exporter := New(collector)
		assert.Implements(t, (*prometheus.Collector)(nil), exporter)
	})
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
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

		exporter := New(collector.New(&http.Client{Transport: roundTripper}))

		ch := make(chan prometheus.Metric, len(mockHosts)*len(mockConnectors)*5+len(mockHosts))
		go func() {
			exporter.Collect(ch)
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

		exporter := New(collector.New(&http.Client{Transport: roundTripper}))

		ch := make(chan prometheus.Metric, 10)
		go func() {
			exporter.Collect(ch)
			close(ch)
		}()

		collected := false
		for range ch {
			collected = true
		}

		assert.False(t, collected)
	})
}
