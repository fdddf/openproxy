package dto

// Stats is a compact view of system statistics for the dashboard.
type Stats struct {
	TotalRequests   int64   `json:"totalRequests"`
	ActiveAPIKeys   int64   `json:"activeAPIKeys"`
	ActiveProviders int64   `json:"activeProviders"`
	SuccessRate     float64 `json:"successRate"`
}
