package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Ecube-labs/kafka-connect-exporter/internal/config"
	applicationError "github.com/Ecube-labs/kafka-connect-exporter/pkg/application-error"
	"github.com/Ecube-labs/kafka-connect-exporter/pkg/logger"
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
		descUnassigned:     prometheus.NewDesc(prefix+"_unassigned", "Kafka Connect Connector Unassigned Task Count", labels, nil),
		descRunning:        prometheus.NewDesc(prefix+"_running", "Kafka Connect Connector Running Task Count", labels, nil),
		descFailed:         prometheus.NewDesc(prefix+"_failed", "Kafka Connect Connector Failed Task Count", labels, nil),
		descPaused:         prometheus.NewDesc(prefix+"_paused", "Kafka Connect Connector Paused Task Count", labels, nil),
		descTaskCount:      prometheus.NewDesc(prefix+"_task_total", "Kafka Connect Connector Total Task Count", labels, nil),
		descConnectorCount: prometheus.NewDesc(prefix+"_total", "Kafka Connect Connector Total Count", []string{"host"}, nil),
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}

	var connectors []string
	if err := json.Unmarshal(body, &connectors); err != nil {
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}

	var status connectorStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, applicationError.New(http.StatusInternalServerError, err.Error(), "")
	}

	return &status, nil
}

// Describe is a no-op, because the collector dynamically allocates metrics.
// https://github.com/prometheus/client_golang/blob/v1.9.0/prometheus/collector.go#L28-L40
func (c *collector) Describe(_ chan<- *prometheus.Desc) {}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	// 각 호스트마다 병렬로 데이터 수집
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
