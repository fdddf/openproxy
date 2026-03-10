package common

//
// ===== Request =====
//

type AIRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
}

// ChatCompletionRequest represents the POST /v1/chat/completions request body.
type ChatCompletionRequest struct {
	Model             string                 `json:"model"`
	Input             any                    `json:"input,omitempty"`               // openai response api
	Instructions      any                    `json:"instructions,omitempty"`        // openai response api
	ParallelToolCalls bool                   `json:"parallel_tool_calls,omitempty"` // openai response api
	Messages          []ChatMessage          `json:"messages"`
	Tools             []ChatCompletionTool   `json:"tools,omitempty"`
	ToolChoice        string                 `json:"tool_choice,omitempty"` // "auto" | "none" | {"type":"function", "function":{"name":"..."}}
	Temperature       float32                `json:"temperature,omitempty"`
	TopP              float32                `json:"top_p,omitempty"`
	Stream            bool                   `json:"stream,omitempty"`
	StreamOptions     StreamOptions          `json:"stream_options,omitempty"`
	MaxTokens         int                    `json:"max_tokens,omitempty"`
	Stop              []string               `json:"stop,omitempty"`
	N                 int                    `json:"n,omitempty"`
	User              string                 `json:"user,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"` // optional, may contain Xcode context
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// ChatMessage represents a message in the conversation.
type ChatMessage struct {
	Role         string            `json:"role"` // "system" | "user" | "assistant" | "tool"
	Content      interface{}       `json:"content,omitempty"`
	ToolCalls    []ChatToolCall    `json:"tool_calls,omitempty"`
	ToolCallID   string            `json:"tool_call_id,omitempty"` // when role == "tool"
	Name         string            `json:"name,omitempty"`
	FunctionCall *ChatFunctionCall `json:"function_call,omitempty"` // legacy OpenAI-style
}

// ChatFunctionCall represents an older style of function call.
type ChatFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

//
// ===== Tools (for Allow Tools) =====
//

// ChatCompletionTool defines a tool (function) that the model can invoke.
type ChatCompletionTool struct {
	Type     string              `json:"type"` // "function"
	Function ChatToolFunctionDef `json:"function"`
}

// ChatToolFunctionDef describes a function the model can call.
type ChatToolFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // JSON Schema object
}

//
// ===== Response =====
//

// ChatCompletionResponse represents the model's reply.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"` // "chat.completion"
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *ChatCompletionUsage   `json:"usage,omitempty"`
}

// ChatCompletionChoice represents a single choice in the response.
type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason,omitempty"` // "stop" | "tool_calls" | "length"
	LogProbs     interface{} `json:"logprobs,omitempty"`
}

// ChatToolCall is returned by the model to indicate tool invocation.
type ChatToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"` // "function"
	Function ChatToolCallFunction `json:"function"`
}

// ChatToolCallFunction specifies which function to call and with what arguments.
type ChatToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string, e.g. {"target":"MyAppTests"}
}

// ChatCompletionUsage optionally includes token statistics.
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo represents a model information
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ListResponse represents a list response
type ListResponse struct {
	Object string      `json:"object"`
	Data   interface{} `json:"data"`
}
