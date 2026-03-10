package iflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type IFlowProvider struct {
	Conf  *common.ProviderConfig
	Model string
}

func (p *IFlowProvider) SetModel(model string) {
	p.Model = model
}

func (p *IFlowProvider) GetApiKey() string {
	return p.Conf.ApiKey
}

func (p *IFlowProvider) RequestMethod() string {
	return "POST"
}

func (p *IFlowProvider) GetChatCompletionURL(ctx *fiber.Ctx) string {
	return p.Conf.BaseURL + "/chat/completions"
}

func (p *IFlowProvider) PrepareRequestBody(ctx *fiber.Ctx, request common.AIRequest) ([]byte, error) {
	var stdReq common.ChatCompletionRequest
	if err := ctx.BodyParser(&stdReq); err != nil {
		log.Debugf("parse body failed: %s, err: %s", ctx.Body(), err)
		return nil, err
	}
	stdReq.Model = p.Model
	return json.Marshal(stdReq)
}

func (p *IFlowProvider) SendRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Conf.ApiKey))
	return pkg.SendHTTPRequest(p.Conf.ProxyURL, req)
}

func (p *IFlowProvider) Response(ctx *fiber.Ctx, resp *http.Response) error {
	defer resp.Body.Close()
	ctx.Set("Content-Type", "application/json")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("iflow: read response body failed: %s", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("read response body failed: %s", err)})
	}
	log.Debugf("iflow: response status=%d, body=%s", resp.StatusCode, string(body))
	return ctx.Send(body)
}

func (p *IFlowProvider) StreamResponse(ctx *fiber.Ctx, resp *http.Response) error {
	ctx.Set("Content-Type", "text/event-stream;charset=utf-8")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")

	ctx.Status(resp.StatusCode)
	encoding := strings.ToLower(resp.Header.Get("Content-Encoding"))
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			w.Flush()
			resp.Body.Close()
		}()

		reader, cleanup := pkg.BuildDecompressReader(encoding, resp)
		if cleanup != nil {
			defer cleanup()
		}

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

		sentFinishChunk := false
		sawDone := false

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, ":") {
				// comment heartbeat from upstream, skip
				log.Debugf("[stream][iflow] comment heartbeat: %s", line)
				continue
			}
			if strings.HasPrefix(strings.ToLower(line), "data:") {
				line = strings.TrimSpace(line[5:])
			}
			if line == "" {
				continue
			}
			if line == "[DONE]" {
				sawDone = true
				break
			}

			raw := []byte(line)
			var chunk StreamChunk
			parseErr := json.Unmarshal(raw, &chunk)
			if parseErr != nil || len(chunk.Choices) == 0 {
				var wrapper StreamChunkWrapper
				if wrapErr := json.Unmarshal(raw, &wrapper); wrapErr != nil {
					var reason error
					if parseErr != nil {
						reason = parseErr
					} else {
						reason = fmt.Errorf("chunk missing choices")
					}
					log.Errorf("[stream][iflow] failed to parse chunk: %v", reason)
					continue
				}

				payload := wrapper.Data
				if len(payload) == 0 {
					payload = wrapper.Result
				}
				if len(payload) == 0 {
					log.Warnf("[stream][iflow] wrapped chunk missing payload, raw: %s", string(raw))
					continue
				}
				if err := json.Unmarshal(payload, &chunk); err != nil {
					log.Errorf("[stream][iflow] failed to parse wrapped chunk: %v, raw: %s", err, string(raw))
					continue
				}
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			if p.applyChunkDefaults(&chunk) {
				sentFinishChunk = true
			}

			payload, marshalErr := json.Marshal(chunk)
			if marshalErr != nil {
				log.Errorf("[stream][iflow] failed to marshal chunk: %v", marshalErr)
				continue
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				log.Errorf("[stream][iflow] error writing to client: %v", err)
				return
			}
			if err := w.Flush(); err != nil {
				log.Errorf("[stream][iflow] error flushing buffer: %v", err)
				return
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			log.Errorf("[stream][iflow] upstream read error: %v", err)
			return
		}

		if !sentFinishChunk {
			if finalChunk := p.buildFinalChunk(); finalChunk != "" {
				if _, err := fmt.Fprintf(w, "data: %s\n\n", finalChunk); err != nil {
					log.Errorf("[stream][iflow] error writing final chunk: %v", err)
					return
				}
			}
		}

		if !sawDone {
			if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
				log.Errorf("[stream][iflow] error writing done signal: %v", err)
				return
			}
		}

		if err := w.Flush(); err != nil {
			log.Errorf("[stream][iflow] error flushing trailer: %v", err)
		}
	})

	return nil
}

func (p *IFlowProvider) buildFinalChunk() string {
	chunk := StreamChunk{
		ID:      "chat",
		Object:  "chat.completion.chunk",
		Created: pkg.GetTimestamp(),
		Model:   p.Model,
		Choices: []StreamChoice{
			{
				Index:        0,
				Delta:        StreamDelta{},
				FinishReason: "stop",
			},
		},
	}
	b, err := json.Marshal(chunk)
	if err != nil {
		return ""
	}
	return string(b)
}

func (p *IFlowProvider) applyChunkDefaults(chunk *StreamChunk) bool {
	if chunk.ID == "" {
		chunk.ID = fmt.Sprintf("chatcmpl-%d", pkg.GetTimestamp())
	}
	if chunk.Object == "" {
		chunk.Object = "chat.completion.chunk"
	}
	if chunk.Model == "" {
		chunk.Model = p.Model
	}
	if chunk.Created == 0 {
		chunk.Created = pkg.GetTimestamp()
	}

	hasFinish := false
	for i := range chunk.Choices {
		delta := &chunk.Choices[i].Delta
		hasPayload := delta.Content != "" ||
			len(delta.Thinking) > 0 ||
			len(delta.ReasoningContent) > 0 ||
			len(delta.ToolCalls) > 0 ||
			delta.FunctionCall != nil ||
			delta.Refusal != ""

		if delta.Role == "" && hasPayload {
			delta.Role = "assistant"
		}
		if chunk.Choices[i].FinishReason != "" {
			hasFinish = true
		}
	}

	return hasFinish
}
