package codex

import (
	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/statics"
)

type ResponseItem map[string]any

type ResponsesRequest struct {
	Model             string          `json:"model"`
	Instructions      string          `json:"instructions"`
	Input             []ResponseItem  `json:"input"`
	Tools             []ResponsesTool `json:"tools,omitempty"`
	Stream            bool            `json:"stream"`
	ToolChoice        string          `json:"tool_choice"`
	Store             bool            `json:"store"`
	ParallelToolCalls bool            `json:"parallel_tool_calls,omitempty"`
	PromptCacheKey    string          `json:"prompt_cache_key,omitempty"`
	Reasoning         any             `json:"reasoning,omitempty"`
	Include           any             `json:"include,omitempty"`
}

type ResponsesTool struct {
	Type        string                 `json:"type"` // "function"
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

func ChatToResponses(req common.ChatCompletionRequest) ResponsesRequest {
	var instructions string
	var input []ResponseItem

	instructions = statics.Prompt

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			continue
		}

		if instructions == "" {
			instructions = msg.Content.(string)
			continue
		}
		t := "input_text"
		if msg.Role == "assistant" {
			t = "output_text"
		}

		input = append(input, ResponseItem{
			"type": "message",
			"role": msg.Role,
			"content": []map[string]string{
				{
					"type": t,
					"text": msg.Content.(string),
				},
			},
		})
	}

	return ResponsesRequest{
		Model:        req.Model,
		Instructions: instructions,
		Input:        input,
		Tools:        ConvertChatToolsToResponses(req.Tools),
		Stream:       req.Stream,
		ToolChoice:   "auto",
	}
}

func ConvertChatToolsToResponses(tools []common.ChatCompletionTool) []ResponsesTool {
	out := make([]ResponsesTool, 0, len(tools))

	for _, t := range tools {
		if t.Type != "function" {
			continue
		}

		out = append(out, ResponsesTool{
			Type:        "function",
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  t.Function.Parameters,
		})
	}

	return out
}
