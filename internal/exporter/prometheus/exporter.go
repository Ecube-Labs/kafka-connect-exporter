package exporter

import (
	"net/http"
	"sync"

	"github.com/Ecube-Labs/kafka-connect-exporter/internal/collector"
	"github.com/Ecube-Labs/kafka-connect-exporter/internal/config"
	applicationError "github.com/Ecube-Labs/kafka-connect-exporter/pkg/application-error"
	"github.com/Ecube-Labs/kafka-connect-exporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// A struct that implements the prometheus.Collector interface.
// https://github.com/prometheus/client_golang/blob/7b39d0144166aa94cc8ce4125bcb3b0da89aad5e/prometheus/collector.go#L27
type exporter struct {
	collector          *collector.Collector
	descUnassigned     *prometheus.Desc
	descRunning        *prometheus.Desc
	descFailed         *prometheus.Desc
	descPaused         *prometheus.Desc
	descTaskCount      *prometheus.Desc
	descConnectorCount *prometheus.Desc
}

func New(collector *collector.Collector) *exporter {
	labels := []string{"connector", "host"}
	prefix := "kafka_connect_connector"

	exporter := &exporter{
		collector:          collector,
		descRunning:        prometheus.NewDesc(prefix+"_running_total", "Total number of tasks in the `RUNNING` state", labels, nil),
		descFailed:         prometheus.NewDesc(prefix+"_failed_total", "Total number of tasks in the `FAILED` state (e.g., due to exceptions reported in status)", labels, nil),
		descPaused:         prometheus.NewDesc(prefix+"_paused_total", "Total number of paused tasks for the Kafka Connect connector", labels, nil),
		descUnassigned:     prometheus.NewDesc(prefix+"_unassigned_total", "Total number of tasks in the `Unassigned` state (i.e., not assigned to any worker.)", labels, nil),
		descTaskCount:      prometheus.NewDesc(prefix+"_task_total", "Total number of tasks for the Kafka Connect connector", labels, nil),
		descConnectorCount: prometheus.NewDesc(prefix+"_total", "Total number of tasks for the connector.", []string{"host"}, nil),
	}
	prometheus.MustRegister(exporter)

	return exporter
}

// Describe is a no-op, because the Collector dynamically allocates metrics.
// https://github.com/prometheus/client_golang/blob/v1.9.0/prometheus/Collector.go#L28-L40
func (e *exporter) Describe(ch chan<- *prometheus.Desc) {}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	// collect metrics for each kafka connect host
	for _, host := range config.KafkaConnectHosts {
		wg.Add(1)

		go func(h string) {
			defer wg.Done()

			connectors, err := e.collector.GetConnectors(h)
			if err != nil {
				logger.Log("error", applicationError.UnWrap(err).Stack)
				return
			}
			ch <- prometheus.MustNewConstMetric(e.descConnectorCount, prometheus.GaugeValue, float64(len(connectors)), h)

			for _, connector := range connectors {
				status, err := e.collector.GetConnectorStatus(h, connector)
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
					e.descUnassigned,
					prometheus.GaugeValue,
					float64(metric.UnAssignedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					e.descRunning,
					prometheus.GaugeValue,
					float64(metric.RunningTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					e.descFailed,
					prometheus.GaugeValue,
					float64(metric.FailedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					e.descPaused,
					prometheus.GaugeValue,
					float64(metric.PausedTaskCount),
					connector, h,
				)

				ch <- prometheus.MustNewConstMetric(
					e.descTaskCount,
					prometheus.GaugeValue,
					float64(metric.TotalTaskCount),
					connector, h,
				)
			}
		}(host)
	}

	wg.Wait()
}

func (e *exporter) Handler() http.Handler {
	return promhttp.Handler()
}
