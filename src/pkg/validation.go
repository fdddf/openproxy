package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Validator provides validation functions
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateAPIKey validates an API key format
func (v *Validator) ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return &ValidationError{Field: "api_key", Message: "API key is required"}
	}

	// Basic validation: API key should be non-empty and not contain spaces
	if strings.Contains(apiKey, " ") {
		return &ValidationError{Field: "api_key", Message: "API key cannot contain spaces"}
	}

	// Check length (basic check - in a real implementation this may vary)
	if len(apiKey) < 10 {
		return &ValidationError{Field: "api_key", Message: "API key is too short"}
	}

	return nil
}

// ValidateModelName validates a model name
func (v *Validator) ValidateModelName(modelName string) error {
	if modelName == "" {
		return &ValidationError{Field: "model", Message: "Model name is required"}
	}

	// Model name should contain only alphanumeric characters, hyphens, and dots
	matched, err := regexp.MatchString(`^[a-zA-Z0-9\-_.]+$`, modelName)
	if err != nil {
		return &ValidationError{Field: "model", Message: "Failed to validate model name"}
	}
	if !matched {
		return &ValidationError{Field: "model", Message: "Model name contains invalid characters"}
	}

	return nil
}

// ValidateMessages validates the messages array in a chat request
func (v *Validator) ValidateMessages(messages []interface{}) error {
	if len(messages) == 0 {
		return &ValidationError{Field: "messages", Message: "Messages array cannot be empty"}
	}

	for i, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			return &ValidationError{Field: fmt.Sprintf("messages[%d]", i), Message: "Message must be an object"}
		}

		role, ok := msgMap["role"].(string)
		if !ok || role == "" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].role", i), Message: "Message role is required"}
		}

		// Validate role is one of the allowed values
		if role != "system" && role != "user" && role != "assistant" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].role", i), Message: "Message role must be 'system', 'user', or 'assistant'"}
		}

		content, ok := msgMap["content"].(string)
		if !ok || content == "" {
			return &ValidationError{Field: fmt.Sprintf("messages[%d].content", i), Message: "Message content is required"}
		}

		// Validate content length (not too long to prevent abuse)
		if len(content) > 10000 { // 10k characters limit
			return &ValidationError{Field: fmt.Sprintf("messages[%d].content", i), Message: "Message content is too long"}
		}
	}

	return nil
}

// ValidateChatCompletionRequest validates a chat completion request body
func (v *Validator) ValidateChatCompletionRequest(request map[string]interface{}) error {
	// Validate model
	model, ok := request["model"].(string)
	if !ok || model == "" {
		return &ValidationError{Field: "model", Message: "Model is required"}
	}
	if err := v.ValidateModelName(model); err != nil {
		return err
	}

	// Validate messages
	messages, ok := request["messages"].([]interface{})
	if !ok {
		return &ValidationError{Field: "messages", Message: "Messages must be an array"}
	}
	if err := v.ValidateMessages(messages); err != nil {
		return err
	}

	// Validate optional fields if present
	if maxTokens, ok := request["max_tokens"].(float64); ok {
		if maxTokens <= 0 {
			return &ValidationError{Field: "max_tokens", Message: "Max tokens must be greater than 0"}
		}
		if maxTokens > 10000 {
			return &ValidationError{Field: "max_tokens", Message: "Max tokens is too large"}
		}
	}

	if temperature, ok := request["temperature"].(float64); ok {
		if temperature < 0 || temperature > 2 {
			return &ValidationError{Field: "temperature", Message: "Temperature must be between 0 and 2"}
		}
	}

	if topP, ok := request["top_p"].(float64); ok {
		if topP < 0 || topP > 1 {
			return &ValidationError{Field: "top_p", Message: "TopP must be between 0 and 1"}
		}
	}

	return nil
}

// SanitizeText sanitizes text input to prevent injection attacks
func (v *Validator) SanitizeText(text string) string {
	// Remove null bytes to prevent some injection attacks
	text = strings.ReplaceAll(text, "\x00", "")

	// Remove control characters (except common whitespace)
	controlCharRegex := regexp.MustCompile(`[\x00-\x1f\x7f]`)
	text = controlCharRegex.ReplaceAllString(text, "")

	return text
}

// SanitizeJSON sanitizes a JSON object to prevent injection
func (v *Validator) SanitizeJSON(data map[string]interface{}) map[string]interface{} {
	for key, value := range data {
		switch val := value.(type) {
		case string:
			data[key] = v.SanitizeText(val)
		case map[string]interface{}:
			data[key] = v.SanitizeJSON(val)
		case []interface{}:
			for i, item := range val {
				if itemStr, ok := item.(string); ok {
					val[i] = v.SanitizeText(itemStr)
				} else if itemMap, ok := item.(map[string]interface{}); ok {
					val[i] = v.SanitizeJSON(itemMap)
				}
			}
		}
	}
	return data
}

// ValidateEmail validates an email address format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "Email is required"}
	}

	// Basic email regex
	matched, err := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	if err != nil {
		return &ValidationError{Field: "email", Message: "Failed to validate email"}
	}
	if !matched {
		return &ValidationError{Field: "email", Message: "Invalid email format"}
	}

	return nil
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) error {
	if username == "" {
		return &ValidationError{Field: "username", Message: "Username is required"}
	}

	// Username should be 3-30 characters, alphanumeric and underscores only
	if len(username) < 3 || len(username) > 30 {
		return &ValidationError{Field: "username", Message: "Username must be 3-30 characters long"}
	}

	matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if err != nil {
		return &ValidationError{Field: "username", Message: "Failed to validate username"}
	}
	if !matched {
		return &ValidationError{Field: "username", Message: "Username can only contain letters, numbers, and underscores"}
	}

	return nil
}
