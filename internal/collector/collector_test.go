package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Should return a new collector", func(t *testing.T) {
		collector := New(http.DefaultClient)
		assert.NotNil(t, collector)
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

		collector := New(&http.Client{Transport: roundTripper})

		connectors, err := collector.GetConnectors("test")
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

		collector := New(&http.Client{Transport: roundTripper})

		connectors, err := collector.GetConnectors("test")
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

		collector := New(&http.Client{Transport: roundTripper})

		status, err := collector.GetConnectorStatus("test", "connector1")
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
		collector := New(&http.Client{Transport: roundTripper})

		status, err := collector.GetConnectorStatus("test", "connector1")
		assert.Nil(t, status)
		assert.NotNil(t, err)
		assert.Equal(t, `Failed to get connector status. status: 444, body: "Internal Server Error"`, err.Error())
	})
}
