package models

// Model model (for model management)
type Model struct {
	CommonFields

	Name       string
	ProviderID uint
	RealModel  string
	IsActive   bool `gorm:"default:true"`
}
