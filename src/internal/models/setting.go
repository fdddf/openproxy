package models

// Setting stores simple key-value configuration persisted in the database.
type Setting struct {
	CommonFields

	Key   string `gorm:"uniqueIndex"`
	Value string
}
