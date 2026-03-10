package iflow

import (
	"encoding/json"

	"github.com/fdddf/openproxy/common"
)

// StreamChoice represents a single streaming choice.
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
	Logprobs     interface{} `json:"logprobs,omitempty"`
}

// StreamDelta matches OpenAI's streaming delta payload.
type StreamDelta struct {
	Role             string           `json:"role,omitempty"`
	Content          string           `json:"content,omitempty"`
	Refusal          string           `json:"refusal,omitempty"`
	Thinking         json.RawMessage  `json:"thinking,omitempty"`
	ReasoningContent json.RawMessage  `json:"reasoning_content,omitempty"`
	ToolCalls        []StreamToolCall `json:"tool_calls,omitempty"`
	// Legacy OpenAI function_call field (still emitted by some SDKs)
	FunctionCall *StreamFunctionCall `json:"function_call,omitempty"`
}

// StreamChunk is an OpenAI-compatible streaming event wrapper.
type StreamChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []StreamChoice              `json:"choices"`
	Usage   *common.ChatCompletionUsage `json:"usage,omitempty"`
}

// StreamChunkWrapper catches non-standard streaming envelopes sent by iflow.
type StreamChunkWrapper struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Result  json.RawMessage `json:"result"`
}

// StreamFunctionCall mirrors the legacy function_call delta shape.
type StreamFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// StreamToolCall matches OpenAI's streaming tool/function call delta.
type StreamToolCall struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function StreamToolFunction `json:"function,omitempty"`
}

// StreamToolFunction holds function call details inside a tool call.
type StreamToolFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}
