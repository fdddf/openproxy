package common

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AIProvider defines the interface for individual AI providers
type AIProvider interface {
	SetModel(model string)
	GetApiKey() string
	PrepareRequestBody(*fiber.Ctx, AIRequest) ([]byte, error)
	GetChatCompletionURL(*fiber.Ctx) string
	SendRequest(req *http.Request) (*http.Response, error)
	RequestMethod() string
	Response(ctx *fiber.Ctx, resp *http.Response) error
	StreamResponse(ctx *fiber.Ctx, resp *http.Response) error
}
