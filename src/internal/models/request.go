package models

// Request model
type Request struct {
	CommonFields

	Method           string
	Path             string
	RequestHeaders   string
	RequestBody      string
	ResponseHeaders  string
	ResponseBody     string
	StatusCode       int
	ResponseTime     int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
	UserID           *uint
	APIKeyID         *uint
	ProviderID       *uint
	ModelID          *uint
}
