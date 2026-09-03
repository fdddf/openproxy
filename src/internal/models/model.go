package models

// Model model (for model management)
type Model struct {
	CommonFields

	Name       string
	ProviderID uint
	RealModel  string
	// No gorm "default" tag: gorm/gen's Save is an upsert-Create, and GORM
	// drops a defaulted field from the INSERT column list when it holds its
	// zero value. With a default tag, IsActive=false was silently discarded on
	// both create and update, so a mapping could never be disabled. The default
	// for new mappings is applied by the handler instead.
	IsActive bool
}
