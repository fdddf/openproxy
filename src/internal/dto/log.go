package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// Log captures the minimal fields needed from system logs.
type Log struct {
	ID        uint      `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// NewLog builds a Log DTO from the database model.
func NewLog(log *models.Log) Log {
	return Log{
		ID:        log.ID,
		Level:     log.Level,
		Message:   log.Message,
		Details:   "",
		Timestamp: log.CreatedAt,
	}
}
