package models

// Log model
type Log struct {
	CommonFields

	Level   string
	Message string
}
