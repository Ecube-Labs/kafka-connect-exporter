package collector

type connectorStatus struct {
	Name      string `json:"name"`
	Connector struct {
		State    string `json:"state"`
		WorkerID string `json:"worker_id"`
	} `json:"connector"`
	Tasks []taskStatus `json:"tasks"`
}

type taskStatus struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
	Trace    string `json:"trace,omitempty"`
}

type ConnectorStatusMetric struct {
	Name                string
	UnAssignedTaskCount int
	RunningTaskCount    int
	PausedTaskCount     int
	FailedTaskCount     int
	TotalTaskCount      int
}
