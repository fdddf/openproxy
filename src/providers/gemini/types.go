package gemini

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content CandidateContent `json:"content"`
}

type CandidateContent struct {
	Parts []Part `json:"parts"`
}

// Types for streaming response
type GeminiStreamEvent struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type OpenAIChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type OpenAIDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIStreamEvent struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}
