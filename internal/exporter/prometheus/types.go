package exporter

type ConnectorStatusMetric struct {
	Name                string
	UnAssignedTaskCount int
	RunningTaskCount    int
	PausedTaskCount     int
	FailedTaskCount     int
	TotalTaskCount      int
}
