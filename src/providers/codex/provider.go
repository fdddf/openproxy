package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/oauth2"
)

const codexRedirectURL = "http://localhost:1455/auth/callback"

type CodexProvider struct {
	Conf  *common.ProviderConfig
	Model string
	OAuth *oauth2.Config
}

func (p *CodexProvider) SetModel(model string) {
	p.Model = model
}

func (p *CodexProvider) GetApiKey() string {
	// For OAuth2 providers, we need to use access token instead of API key
	return p.Conf.AccessToken
}

func (p *CodexProvider) RequestMethod() string {
	return "POST"
}

func (p *CodexProvider) GetChatCompletionURL(ctx *fiber.Ctx) string {
	return "https://chatgpt.com/backend-api/codex/responses"
}

func (p *CodexProvider) PrepareRequestBody(ctx *fiber.Ctx, request common.AIRequest) ([]byte, error) {
	isChatCompletion := ctx.Path() == "/v1/chat/completions"
	if isChatCompletion {
		// convert client request to codex responses request
		var stdReq common.ChatCompletionRequest
		if err := ctx.BodyParser(&stdReq); err != nil {
			log.Debugf("parse body failed: %s, err: %s", ctx.Body(), err)
			return nil, err
		}
		stdReq.Model = p.Model
		responses := ChatToResponses(stdReq)
		return json.Marshal(responses)
	}

	body := ctx.Body()
	var aReq map[string]any
	if err := json.Unmarshal(body, &aReq); err != nil {
		log.Debugf("parse body failed: %s, err: %s", body, err)
		return nil, err
	}
	aReq["model"] = p.Model
	return json.Marshal(aReq)
}

func (p *CodexProvider) SendRequest(req *http.Request) (*http.Response, error) {
	// Check if token needs refresh
	if !p.Conf.TokenExpiry.IsZero() && time.Now().After(p.Conf.TokenExpiry) {
		// Refresh the token
		token := &oauth2.Token{
			AccessToken:  p.Conf.AccessToken,
			RefreshToken: p.Conf.RefreshToken,
			Expiry:       p.Conf.TokenExpiry,
		}

		newToken, err := p.OAuth.TokenSource(context.Background(), token).Token()
		if err == nil {
			// Update the config with new token
			p.Conf.AccessToken = newToken.AccessToken
			p.Conf.RefreshToken = newToken.RefreshToken

			// Update expiry time
			p.Conf.TokenExpiry = newToken.Expiry
		} else {
			log.Errorf("Failed to refresh token: %v", err)
		}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Conf.AccessToken))
	req.Header.Set("ChatGPT-Account-ID", p.Conf.AccountID)
	req.Header.Set("Host", "chatgpt.com")
	req.Header.Set("Originator", "codex_vscode")
	req.Header.Del("X-Title")
	req.Header.Del("Http-Referer")

	// if user-agent doesn't contains codex, rewrite it
	if !strings.Contains(req.Header.Get("User-Agent"), "codex") {
		req.Header.Set("User-Agent", "codex_vscode/0.70.0-alpha.4 (Windows 10.0.26100; x86_64) unknown (VS Code; 0.4.51)")
	}
	log.Debugf("[proxy] request headers %+v", req.Header)

	return pkg.SendHTTPRequest(p.Conf.ProxyURL, req)
}

func (p *CodexProvider) Response(ctx *fiber.Ctx, resp *http.Response) error {
	defer resp.Body.Close()
	ctx.Set("Content-Type", "application/json")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("codex: read response body failed: %s", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("read response body failed: %s", err)})
	}
	log.Debugf("codex: response status=%d, body=%s", resp.StatusCode, string(body))
	return ctx.Send(body)
}

func (p *CodexProvider) StreamResponse(ctx *fiber.Ctx, resp *http.Response) error {
	ctx.Set("Content-Type", "text/event-stream;charset=utf-8")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")

	ctx.Status(resp.StatusCode)
	encoding := strings.ToLower(resp.Header.Get("Content-Encoding"))
	isChatCompletion := ctx.Path() == "/v1/chat/completions"
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			w.Flush()
			resp.Body.Close()
		}()

		reader, cleanup := pkg.BuildDecompressReader(encoding, resp)
		if cleanup != nil {
			defer cleanup()
		}

		if isChatCompletion {
			if streamCodexChatCompletion(reader, w, p.Model) {
				return
			}
			return
		}

		buf := make([]byte, 32*1024)
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

// GetAuthCodeURL returns the OAuth2 authorization URL for the Codex provider
func (p *CodexProvider) GetAuthCodeURL(state string) string {
	return p.OAuth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *CodexProvider) ExchangeCode(ctx context.Context, code string, codeVerifier string) (*oauth2.Token, error) {
	// Custom exchange to avoid OpenAI quirks with PKCE and client secret handling.
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", codexRedirectURL)
	values.Set("client_id", p.Conf.ClientID)
	if codeVerifier != "" {
		values.Set("code_verifier", codeVerifier)
	}
	// Only send client_secret when configured; Codex public client omits it.
	if strings.TrimSpace(p.Conf.ClientSecret) != "" {
		values.Set("client_secret", p.Conf.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.OAuth.Endpoint.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	log.Debugf("[oauth] request: %s %s, with proxy", req.Method, req.URL, p.Conf.ProxyURL)
	resp, err := pkg.SendHTTPRequest(p.Conf.ProxyURL, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		dump, _ := httputil.DumpResponse(resp, false)
		return nil, fmt.Errorf("token exchange failed: status=%d body=%s resp=%s", resp.StatusCode, string(body), string(dump))
	}
	log.Debugf("[oauth] response: %s", string(body))

	raw := map[string]interface{}{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  getStringField(raw, "access_token"),
		TokenType:    getStringField(raw, "token_type"),
		RefreshToken: getStringField(raw, "refresh_token"),
	}

	if expiresIn := getFloatField(raw, "expires_in"); expiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}

	// Preserve the raw fields so callers can read chatgpt_account_id, etc.
	token = token.WithExtra(raw)

	log.Debugf("[oauth] token: %+v", token)
	if token.AccessToken == "" {
		return nil, errors.New("token exchange failed: missing access_token")
	}
	return token, nil
}

// RefreshToken refreshes the access token using the refresh token
func (p *CodexProvider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{RefreshToken: refreshToken}
	return p.OAuth.TokenSource(ctx, token).Token()
}
