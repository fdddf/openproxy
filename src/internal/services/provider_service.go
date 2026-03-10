package services

import (
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/consts"
	"github.com/fdddf/openproxy/providers/antigravity"
	"github.com/fdddf/openproxy/providers/codex"
	"github.com/fdddf/openproxy/providers/gemini"
	"github.com/fdddf/openproxy/providers/iflow"
	"github.com/fdddf/openproxy/providers/openai"
	"golang.org/x/oauth2"
)

// ProviderServiceImpl implements the ProviderService interface
type ProviderServiceImpl struct{}

// NewProviderService creates a new ProviderService instance
func NewProviderService() ProviderService {
	return &ProviderServiceImpl{}
}

// GetModelProvider returns the appropriate provider based on configuration
func (p *ProviderServiceImpl) GetModelProvider(providerConfig *common.ProviderConfig, model string) (common.AIProvider, error) {
	switch providerConfig.Type {
	case consts.ProviderGemini.String():
		return &gemini.GeminiProvider{
			Conf:  providerConfig,
			Model: model,
		}, nil
	case consts.ProviderIflow.String():
		return &iflow.IFlowProvider{
			Conf:  providerConfig,
			Model: model,
		}, nil
	case consts.ProviderCodex.String():
		return &codex.CodexProvider{
			Conf:  providerConfig,
			Model: model,
			OAuth: buildCodexOAuthConfig(providerConfig),
		}, nil
	case consts.ProviderPlatformAntigravity.String():
		return &antigravity.AntigravityProvider{
			Conf:  providerConfig,
			Model: model,
			OAuth: buildAntigravityOAuthConfig(providerConfig),
		}, nil

	default:
		return &openai.OpenAIProvider{
			Conf:  providerConfig,
			Model: model,
		}, nil
	}
}

func buildCodexOAuthConfig(conf *common.ProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  selectRedirect(conf.RedirectURL),
		Scopes:       splitScopes(conf.Scopes),
		Endpoint: oauth2.Endpoint{
			AuthURL:  sanitizeCodexAuthURL(conf.AuthURL),
			TokenURL: sanitizeTokenURL(conf.TokenURL),
		},
	}
}

func buildAntigravityOAuthConfig(conf *common.ProviderConfig) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		RedirectURL:  conf.RedirectURL,
		Scopes:       splitScopes(conf.Scopes),
		Endpoint: oauth2.Endpoint{
			AuthURL:  conf.AuthURL,
			TokenURL: conf.TokenURL,
		},
	}
}

func splitScopes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})

	scopes := make([]string, 0, len(fields))
	for _, scope := range fields {
		if trimmed := strings.TrimSpace(scope); trimmed != "" {
			scopes = append(scopes, trimmed)
		}
	}
	return scopes
}

func selectRedirect(raw string) string {
	if strings.TrimSpace(raw) != "" {
		return raw
	}
	return "http://localhost:1455/auth/callback"
}
