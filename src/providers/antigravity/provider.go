package antigravity

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/oauth2"
)

type AntigravityProvider struct {
	Conf  *common.ProviderConfig
	Model string
	OAuth *oauth2.Config
}

func (p *AntigravityProvider) SetModel(model string) {
	p.Model = model
}

func (p *AntigravityProvider) GetApiKey() string {
	return p.Conf.AccessToken
}

func (p *AntigravityProvider) RequestMethod() string {
	return http.MethodPost
}

func (p *AntigravityProvider) GetChatCompletionURL(ctx *fiber.Ctx) string {
	return p.Conf.BaseURL + "/chat/completions"
}

func (p *AntigravityProvider) PrepareRequestBody(ctx *fiber.Ctx, request common.AIRequest) ([]byte, error) {
	var stdReq common.ChatCompletionRequest
	if err := ctx.BodyParser(&stdReq); err != nil {
		log.Debugf("parse body failed: %s, err: %s", ctx.Body(), err)
		return nil, err
	}
	stdReq.Model = p.Model
	return json.Marshal(stdReq)
}

func (p *AntigravityProvider) SendRequest(req *http.Request) (*http.Response, error) {
	if !p.Conf.TokenExpiry.IsZero() && time.Now().After(p.Conf.TokenExpiry) && p.OAuth != nil {
		token := &oauth2.Token{
			AccessToken:  p.Conf.AccessToken,
			RefreshToken: p.Conf.RefreshToken,
			Expiry:       p.Conf.TokenExpiry,
		}

		newToken, err := p.OAuth.TokenSource(context.Background(), token).Token()
		if err == nil {
			p.Conf.AccessToken = newToken.AccessToken
			if newToken.RefreshToken != "" {
				p.Conf.RefreshToken = newToken.RefreshToken
			}
			p.Conf.TokenExpiry = newToken.Expiry
		} else {
			log.Errorf("[antigravity oauth] failed to refresh token: %v", err)
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Conf.AccessToken))
	return pkg.SendHTTPRequest(p.Conf.ProxyURL, req)
}

func (p *AntigravityProvider) Response(ctx *fiber.Ctx, resp *http.Response) error {
	defer resp.Body.Close()
	ctx.Set("Content-Type", "application/json")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("antigravity: read response body failed: %s", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("read response body failed: %s", err)})
	}
	log.Debugf("antigravity: response status=%d, body=%s", resp.StatusCode, string(body))
	return ctx.Send(body)
}

func (p *AntigravityProvider) StreamResponse(ctx *fiber.Ctx, resp *http.Response) error {
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

		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					log.Debugf("[stream] error writing to client: %v", writeErr)
					return
				}
				if err := w.Flush(); err != nil {
					log.Debugf("[stream] error flushing buffer: %v", err)
					return
				}
			}

			if err == io.EOF {
				log.Debug("[stream] upstream closed normally")
				break
			} else if err != nil {
				log.Debugf("[stream] error reading upstream: %v", err)
				return
			}
		}
	})

	return nil
}
