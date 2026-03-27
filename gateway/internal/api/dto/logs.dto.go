package dto

type IngestLogsRequest struct {
	Logs []string `json:"logs" binding:"required"`
}
