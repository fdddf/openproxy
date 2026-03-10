package services

import (
	"context"
	"errors"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/oauth2"
)

// https://github.com/linwanxiaoyehua/AntiProxy/blob/main/src/modules/oauth.rs

type antigravityOAuthService struct {
}

func newAntigravityOAuthService() antigravityOAuthService {
	return antigravityOAuthService{}
}

func (s antigravityOAuthService) logExchangeError(err error) {
	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		log.Errorf("[anti oauth] code exchange failed: status=%d body=%s", retrieveErr.Response.StatusCode, string(retrieveErr.Body))
		return
	}
	log.Errorf("[anti oauth] code exchange failed: %v", err)
}

func (s antigravityOAuthService) ExchangeCode(ctx context.Context, provider *models.Provider, code, verifier string) (*oauth2.Token, error) {
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

	oauthConf := buildAntigravityOAuthConfig(conf)

	options := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_verifier", verifier),
		oauth2.SetAuthURLParam("access_type", "offline"),
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	}

	if trimmed := strings.TrimSpace(provider.RedirectURL); trimmed != "" {
		oauthConf.RedirectURL = trimmed
	}

	token, err := oauthConf.Exchange(ctx, code, options...)
	if err != nil {
		s.logExchangeError(err)
		return nil, err
	}
	return token, nil
}

func (s antigravityOAuthService) ApplyToken(provider *models.Provider, token *oauth2.Token) {
	provider.AccessToken = token.AccessToken
	provider.RefreshToken = token.RefreshToken
	provider.TokenExpiry = &token.Expiry
}
