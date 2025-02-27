package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	applicationError "github.com/Ecube-Labs/kafka-connect-exporter/pkg/application-error"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

// A struct that implements the prometheus.Collector interface.
// https://github.com/prometheus/client_golang/blob/7b39d0144166aa94cc8ce4125bcb3b0da89aad5e/prometheus/collector.go#L27
type collector struct {
	client             *http.Client
	descUnassigned     *prometheus.Desc
	descRunning        *prometheus.Desc
	descFailed         *prometheus.Desc
	descPaused         *prometheus.Desc
	descTaskCount      *prometheus.Desc
	descConnectorCount *prometheus.Desc
}

func New() *collector {
	labels := []string{"connector", "host"}
	prefix := "kafka_connect_connector"

	return &collector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		descRunning:        prometheus.NewDesc(prefix+"_running_total", "Total number of tasks in the `RUNNING` state", labels, nil),
		descFailed:         prometheus.NewDesc(prefix+"_failed_total", "Total number of tasks in the `FAILED` state (e.g., due to exceptions reported in status)", labels, nil),
		descPaused:         prometheus.NewDesc(prefix+"_paused_total", "Total number of paused tasks for the Kafka Connect connector", labels, nil),
		descUnassigned:     prometheus.NewDesc(prefix+"_unassigned_total", "Total number of tasks in the `Unassigned` state (i.e., not assigned to any worker.)", labels, nil),
		descTaskCount:      prometheus.NewDesc(prefix+"_task_total", "Total number of tasks for the Kafka Connect connector", labels, nil),
		descConnectorCount: prometheus.NewDesc(prefix+"_total", "Total number of tasks for the connector.", []string{"host"}, nil),
	}
}

// Get a list of kafka connect connectors from the given host
func (c *collector) getConnectors(host string) ([]string, error) {
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
func (c *collector) getConnectorStatus(host string, connector string) (*connectorStatus, error) {
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

// Describe is a no-op, because the collector dynamically allocates metrics.
// https://github.com/prometheus/client_golang/blob/v1.9.0/prometheus/collector.go#L28-L40
func (c *collector) Describe(_ chan<- *prometheus.Desc) {}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	// collect metrics for each kafka connect host
	for _, host := range config.KafkaConnectHosts {
		wg.Add(1)

		go func(h string) {
			defer wg.Done()

			connectors, err := c.getConnectors(h)
			if err != nil {
				logger.Log("error", applicationError.UnWrap(err).Stack)
				return
			}
			ch <- prometheus.MustNewConstMetric(c.descConnectorCount, prometheus.GaugeValue, float64(len(connectors)), h)

			for _, connector := range connectors {
				status, err := c.getConnectorStatus(h, connector)
				if err != nil {
					logger.Log("error", applicationError.UnWrap(err).Stack)
					continue
				}

				var unAssignedTaskCount int
				var runningTaskCount int
				var pausedTaskCount int
				var failedTaskCount int

				for _, task := range status.Tasks {
					switch task.State {
					case "RUNNING":
						runningTaskCount++
					case "PAUSED":
						pausedTaskCount++
					case "FAILED":
						failedTaskCount++
					default:
						unAssignedTaskCount++
					}
				}
				metric := ConnectorStatusMetric{
					UnAssignedTaskCount: unAssignedTaskCount,
					RunningTaskCount:    runningTaskCount,
					PausedTaskCount:     pausedTaskCount,
					FailedTaskCount:     failedTaskCount,
					TotalTaskCount:      len(status.Tasks),
				}

				ch <- prometheus.MustNewConstMetric(
					c.descUnassigned,
					prometheus.GaugeValue,
					float64(metric.UnAssignedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					c.descRunning,
					prometheus.GaugeValue,
					float64(metric.RunningTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					c.descFailed,
					prometheus.GaugeValue,
					float64(metric.FailedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					c.descPaused,
					prometheus.GaugeValue,
					float64(metric.PausedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					c.descTaskCount,
					prometheus.GaugeValue,
					float64(metric.TotalTaskCount),
					connector, h,
				)
			}
		}(host)
	}

	wg.Wait()
}
