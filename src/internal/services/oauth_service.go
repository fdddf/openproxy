package services

import (
	"context"

	"github.com/fdddf/openproxy/internal/models"
	"golang.org/x/oauth2"
)

type oauthService interface {
	ExchangeCode(ctx context.Context, provider *models.Provider, code, verifier string) (*oauth2.Token, error)
	ApplyToken(provider *models.Provider, token *oauth2.Token)
}
