package codex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/pkg"
	"github.com/gofiber/fiber/v2/log"
)

type responseEvent struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	Model       string `json:"model"`
	Delta       string `json:"delta"`
	OutputText  string `json:"output_text"`
	OutputIndex int    `json:"output_index"`
	Usage       *struct {
		InputTokens      int `json:"input_tokens"`
		OutputTokens     int `json:"output_tokens"`
		TotalTokens      int `json:"total_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type streamDelta struct {
	Content string `json:"content,omitempty"`
}

type streamChoice struct {
	Index        int         `json:"index"`
	Delta        streamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type streamChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []streamChoice              `json:"choices"`
	Usage   *common.ChatCompletionUsage `json:"usage,omitempty"`
}

func streamCodexChatCompletion(reader io.Reader, w *bufio.Writer, model string) bool {
	toUsage := func(src *struct {
		InputTokens      int `json:"input_tokens"`
		OutputTokens     int `json:"output_tokens"`
		TotalTokens      int `json:"total_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	}) *common.ChatCompletionUsage {
		if src == nil {
			return nil
		}

		prompt := src.PromptTokens
		if prompt == 0 && src.InputTokens > 0 {
			prompt = src.InputTokens
		}

		completion := src.CompletionTokens
		if completion == 0 && src.OutputTokens > 0 {
			completion = src.OutputTokens
		}

		total := src.TotalTokens
		if total == 0 && (prompt > 0 || completion > 0) {
			total = prompt + completion
		}

		return &common.ChatCompletionUsage{
			PromptTokens:     prompt,
			CompletionTokens: completion,
			TotalTokens:      total,
		}
	}

	writeChunk := func(chunk streamChunk) bool {
		if chunk.Created == 0 {
			chunk.Created = pkg.GetTimestamp()
		}
		if chunk.Model == "" {
			chunk.Model = model
		}
		raw, err := json.Marshal(chunk)
		if err != nil {
			log.Debugf("[stream][codex] marshal chunk failed: %v", err)
			return false
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", raw); err != nil {
			log.Debugf("[stream][codex] error writing to client: %v", err)
			return false
		}
		if err := w.Flush(); err != nil {
			log.Debugf("[stream][codex] error flushing buffer: %v", err)
			return false
		}
		return true
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var (
		eventName   string
		respID      string
		respModel   string
		respCreated int64
		sentDone    bool
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			eventName = ""
			continue
		}
		if strings.HasPrefix(line, ":") {
			// heartbeat/comment
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "event:") {
			eventName = strings.TrimSpace(line[6:])
			continue
		}
		if !strings.HasPrefix(lower, "data:") {
			log.Debugf("[stream][codex] invalid line: %s", line)
			continue
		}

		payload := strings.TrimSpace(line[5:])
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			sentDone = true
			break
		}

		var evt responseEvent
		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			log.Debugf("[stream][codex] parse chunk failed: %v, raw: %s", err, payload)
			continue
		}

		if evt.Object != "" && eventName == "" {
			eventName = evt.Object
		}

		if evt.ID != "" {
			respID = evt.ID
		}
		if evt.Model != "" {
			respModel = evt.Model
		}
		if evt.Created != 0 {
			respCreated = evt.Created
		}

		switch eventName {
		case "response.output_text.delta":
			chunk := streamChunk{
				ID:      respID,
				Object:  "chat.completion.chunk",
				Created: respCreated,
				Model:   respModel,
				Choices: []streamChoice{
					{
						Index: evt.OutputIndex,
						Delta: streamDelta{Content: evt.Delta},
					},
				},
			}
			if !writeChunk(chunk) {
				return true
			}
		case "response.output_text.done", "response.completed":
			chunk := streamChunk{
				ID:      respID,
				Object:  "chat.completion.chunk",
				Created: respCreated,
				Model:   respModel,
				Choices: []streamChoice{
					{
						Index:        evt.OutputIndex,
						FinishReason: "stop",
					},
				},
				Usage: toUsage(evt.Usage),
			}
			if !writeChunk(chunk) {
				return true
			}
			if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
				log.Debugf("[stream][codex] error writing done signal: %v", err)
				return true
			}
			if err := w.Flush(); err != nil {
				log.Debugf("[stream][codex] error flushing done signal: %v", err)
			}
			return true
		}

		eventName = ""
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Debugf("[stream][codex] upstream read error: %v", err)
	}

	if !sentDone {
		if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
			log.Debugf("[stream][codex] error writing done fallback: %v", err)
			return true
		}
		if err := w.Flush(); err != nil {
			log.Debugf("[stream][codex] error flushing done fallback: %v", err)
		}
	}

	return true
}
