package codex

import (
	"fmt"
	"strconv"
)

func getStringField(raw map[string]interface{}, key string) string {
	if raw == nil {
		return ""
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getFloatField(raw map[string]interface{}, key string) float64 {
	if raw == nil {
		return 0
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
