package models

import "time"

// APIKey model
type APIKey struct {
	CommonFields

	Key       string     `gorm:"unique" json:"key"`
	Name      string     `json:"name"`
	UserID    uint       `json:"user_id"`
	Quota     *int64     `json:"quota,omitempty"`       // Request quota (nil means unlimited)
	Used      int64      `json:"used" gorm:"default:0"` // Requests used
	ResetTime *time.Time `json:"reset_time,omitempty"`  // When quota resets
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
