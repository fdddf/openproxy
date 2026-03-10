package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/consts"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/fdddf/openproxy/providers/codex"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/oauth2"
)

type codexOAuthService struct {
}

func newCodexOAuthService() codexOAuthService {
	return codexOAuthService{}
}

func (s codexOAuthService) ExchangeCode(ctx context.Context, provider *models.Provider, code, verifier string) (*oauth2.Token, error) {
	conf := &common.ProviderConfig{
		Type:         provider.Platform.String(),
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		Scopes:       provider.Scopes,
		ProxyURL:     provider.ProxyURL,
	}

	if provider.Platform == consts.ProviderPlatformCodex {
		conf.AuthURL = sanitizeCodexAuthURL(provider.AuthURL)
		conf.TokenURL = sanitizeTokenURL(provider.TokenURL)
		conf.RedirectURL = selectRedirect(provider.RedirectURL)
	}

	codexProvider := &codex.CodexProvider{
		Conf:  conf,
		OAuth: buildCodexOAuthConfig(conf),
	}

	token, err := codexProvider.ExchangeCode(ctx, code, verifier)
	if err != nil {
		s.logExchangeError(err)
		return nil, err
	}
	return token, nil
}

func (s codexOAuthService) ApplyToken(provider *models.Provider, token *oauth2.Token) {
	provider.AccessToken = token.AccessToken
	provider.RefreshToken = token.RefreshToken
	provider.TokenExpiry = &token.Expiry

	accountID, err := s.parseChatGPTAccountID(token.AccessToken)
	if err != nil {
		log.Warnf("[codex oauth] failed to parse account id from access token: %v", err)
		return
	}
	if trimmed := strings.TrimSpace(accountID); trimmed != "" {
		provider.AccountID = trimmed
	}
}

func (s codexOAuthService) logExchangeError(err error) {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		log.Errorf("[codex oauth] code exchange failed: status=%d body=%s", retrieveErr.Response.StatusCode, string(retrieveErr.Body))
		return
	}
	log.Errorf("[codex oauth] code exchange failed: %v", err)
}

// parseChatGPTAccountID extracts chatgpt_account_id from the JWT access token payload.
func (s codexOAuthService) parseChatGPTAccountID(accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("empty access token")
	}

	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid jwt format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", fmt.Errorf("unmarshal jwt payload: %w", err)
	}
	log.Debugf("[codex oauth] jwt payload: %v", claims)

	authClaim, ok := claims["https://api.openai.com/auth"].(map[string]interface{})
	if !ok || authClaim == nil {
		return "", nil
	}
	log.Debugf("[codex oauth] auth claim: %v", authClaim)

	if val, exists := authClaim["chatgpt_account_id"]; exists && val != nil {
		if s, ok := val.(string); ok {
			return strings.TrimSpace(s), nil
		}
		return strings.TrimSpace(fmt.Sprintf("%v", val)), nil
	}

	return "", nil
}
