package gemini

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

type GeminiProvider struct {
	Conf  *common.ProviderConfig
	Model string
}

func (p *GeminiProvider) SetModel(model string) {
	p.Model = model
}

func (p *GeminiProvider) GetApiKey() string {
	return p.Conf.ApiKey
}

func (p *GeminiProvider) RequestMethod() string {
	return "POST"
}

func (p *GeminiProvider) GetChatCompletionURL(ctx *fiber.Ctx) string {
	return fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.Conf.BaseURL, p.Model, p.Conf.ApiKey)
}

func (p *GeminiProvider) PrepareRequestBody(ctx *fiber.Ctx, request common.AIRequest) ([]byte, error) {
	req := GeminiRequest{
		Contents: make([]Content, 0),
	}

	var stdReq common.ChatCompletionRequest
	if err := ctx.BodyParser(&stdReq); err != nil {
		log.Debugf("parse body failed: %s, err: %s", ctx.Body(), err)
		return nil, err
	}

	for _, message := range stdReq.Messages {
		var contentText string

		switch v := message.Content.(type) {
		case string:
			contentText = v
		case []interface{}:
			// Handle array of content parts - concatenate text parts
			var textParts []string
			for _, item := range v {
				if text, ok := item.(map[string]interface{})["text"].(string); ok {
					textParts = append(textParts, text)
				}
			}
			contentText = strings.Join(textParts, " ")
		case map[string]interface{}:
			// Handle single content part object
			if text, ok := v["text"].(string); ok {
				contentText = text
			} else {
				contentText = fmt.Sprintf("%v", v)
			}
		default:
			contentText = fmt.Sprintf("%v", v)
		}

		req.Contents = append(req.Contents, Content{
			Parts: []Part{
				{
					Text: contentText,
				},
			},
			Role: message.Role,
		})
	}
	return json.Marshal(req)
}

func (p *GeminiProvider) SendRequest(req *http.Request) (*http.Response, error) {
	return pkg.SendHTTPRequest(p.Conf.ProxyURL, req)
}

func (p *GeminiProvider) Response(ctx *fiber.Ctx, resp *http.Response) error {
	defer resp.Body.Close()
	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		log.Errorf("gemini: parse response failed: %s", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("parse gemini response failed: %s", err)})
	}

	if len(geminiResp.Candidates) == 0 {
		log.Errorf("gemini: response is empty, response=%v", geminiResp)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gemini response is empty"})
	}

	to := common.ChatCompletionResponse{
		Choices: []common.ChatCompletionChoice{
			{
				Message: common.ChatMessage{
					Role:    "assistant",
					Content: geminiResp.Candidates[0].Content.Parts[0].Text,
				},
			},
		},
	}

	log.Debugf("gemini: response status=%d, candidates=%d", resp.StatusCode, len(geminiResp.Candidates))
	return ctx.JSON(to)
}

func (p *GeminiProvider) StreamResponse(ctx *fiber.Ctx, resp *http.Response) error {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// Send final event to indicate completion
					finalEvent := OpenAIStreamEvent{
						ID:      "chatcmpl-123",
						Object:  "chat.completion.chunk",
						Created: pkg.GetTimestamp(),
						Model:   p.Model,
						Choices: []OpenAIChoice{
							{
								Index:        0,
								Delta:        OpenAIDelta{},
								FinishReason: "stop",
							},
						},
					}
					data, _ := json.Marshal(finalEvent)
					fmt.Fprintf(w, "data: %s\n\n", data)
					w.Flush()
					break
				}
				return
			}
			lineStr := string(line)
			if strings.HasPrefix(lineStr, "data: ") {
				jsonStr := strings.TrimPrefix(lineStr, "data: ")
				if strings.TrimSpace(jsonStr) == "[DONE]" {
					break
				}

				var geminiEvent GeminiStreamEvent
				if err := json.Unmarshal([]byte(jsonStr), &geminiEvent); err != nil {
					return
				}
				if len(geminiEvent.Candidates) > 0 && len(geminiEvent.Candidates[0].Content.Parts) > 0 {
					text := geminiEvent.Candidates[0].Content.Parts[0].Text
					if text != "" {
						openAIEvent := OpenAIStreamEvent{
							ID:      "chatcmpl-123",
							Object:  "chat.completion.chunk",
							Created: pkg.GetTimestamp(),
							Model:   p.Model,
							Choices: []OpenAIChoice{
								{
									Index: 0,
									Delta: OpenAIDelta{
										Content: text,
										Role:    "assistant",
									},
								},
							},
						}
						data, _ := json.Marshal(openAIEvent)
						fmt.Fprintf(w, "data: %s\n\n", data)
						w.Flush()
					}
				}
			}
		}
	})
	return nil
}
