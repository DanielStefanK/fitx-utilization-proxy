package responses

type UtilizationResponse struct {
	StartTime string      `json:"startTime"`
	EndTime   string      `json:"endTime"`
	Items     []DataPoint `json:"items"`
	UUID      string      `json:"uuid"`
	Workload  uint8       `json:"workload"`
	Name      string      `json:"name"`
}

type DataPoint struct {
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	IsCurrent  bool   `json:"isCurrent"`
	Level      string `json:"level"`
	Percentage uint8  `json:"percentage"`
}
