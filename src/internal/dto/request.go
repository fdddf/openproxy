package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// Request summarizes a request record for frontend consumption.
type Request struct {
	ID               uint      `json:"id"`
	Method           string    `json:"method"`
	Path             string    `json:"path"`
	StatusCode       int       `json:"statusCode"`
	Success          bool      `json:"success"`
	RequestTime      time.Time `json:"requestTime"`
	ResponseTime     int       `json:"responseTime"`
	PromptTokens     int       `json:"promptTokens"`
	CompletionTokens int       `json:"completionTokens"`
	TotalTokens      int       `json:"totalTokens"`
	Cost             float64   `json:"cost"`
	RequestHeaders   string    `json:"requestHeaders,omitempty"`
	RequestBody      string    `json:"requestBody,omitempty"`
	ResponseHeaders  string    `json:"responseHeaders,omitempty"`
	ResponseBody     string    `json:"responseBody,omitempty"`
	ProviderID       *uint     `json:"providerId,omitempty"`
	ModelID          *uint     `json:"modelId,omitempty"`
	APIKeyID         *uint     `json:"apiKeyId,omitempty"`
}

// NewRequest builds a Request DTO from the database model.
func NewRequest(req *models.Request) Request {
	success := req.StatusCode >= 200 && req.StatusCode < 300

	return Request{
		ID:               req.ID,
		Method:           req.Method,
		Path:             req.Path,
		StatusCode:       req.StatusCode,
		Success:          success,
		RequestTime:      req.CreatedAt,
		ResponseTime:     req.ResponseTime,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		TotalTokens:      req.TotalTokens,
		Cost:             req.Cost,
		RequestHeaders:   req.RequestHeaders,
		RequestBody:      req.RequestBody,
		ResponseHeaders:  req.ResponseHeaders,
		ResponseBody:     req.ResponseBody,
		ProviderID:       req.ProviderID,
		ModelID:          req.ModelID,
		APIKeyID:         req.APIKeyID,
	}
}
