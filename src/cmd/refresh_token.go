package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/fdddf/openproxy/internal/services"
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func newRefreshTokenCmd() *cobra.Command {
	var providerID int64
	var all bool

	cmd := &cobra.Command{
		Use:   "refresh-token",
		Short: "Manually refresh OAuth tokens for providers",
		Long: `Manually refresh OAuth tokens for providers.
		
This command allows you to refresh OAuth tokens for specific providers or all providers.
It uses the same logic as the health check service to renew expired tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRefreshToken(configPath, providerID, all)
		},
	}

	cmd.Flags().Int64VarP(&providerID, "provider", "p", 0, "Provider ID to refresh (required if not using --all)")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Refresh tokens for all OAuth providers")

	// Mark provider flag as required if --all is not specified
	cmd.MarkFlagsMutuallyExclusive("provider", "all")

	return cmd
}

func runRefreshToken(cfgPath string, providerID int64, refreshAll bool) error {
	// Initialize services
	configService := services.NewConfigService()
	if err := configService.LoadConfig(cfgPath); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbService := services.NewDatabaseService(configService)
	if err := dbService.InitDatabase(); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	// Get the DAO to query providers
	dao := dbService.GetDAO()

	// Determine which providers to refresh
	var providers []*models.Provider
	if refreshAll {
		// Refresh all OAuth providers (those with RefreshToken set)
		var err error
		providers, err = dao.Provider.Where(dao.Provider.RefreshToken.IsNotNull()).Find()
		if err != nil {
			return fmt.Errorf("load OAuth providers: %w", err)
		}
	} else {
		if providerID == 0 {
			return fmt.Errorf("either provider ID or --all flag must be specified")
		}

		// Refresh specific provider
		provider, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(providerID))).First()
		if err != nil {
			return fmt.Errorf("provider with ID %d not found: %w", providerID, err)
		}

		if provider.RefreshToken == "" {
			return fmt.Errorf("provider %d (%s) does not have a refresh token configured", providerID, provider.Name)
		}

		providers = append(providers, provider)
	}

	// Refresh each provider's token
	for _, provider := range providers {
		if err := refreshProviderToken(dbService, provider); err != nil {
			log.Errorf("Failed to refresh token for provider %d (%s): %v", provider.ID, provider.Name, err)
			continue
		}

		fmt.Printf("Successfully refreshed token for provider %d (%s)\n", provider.ID, provider.Name)
	}

	// Close database connection
	if gormDB := dbService.GetDB(); gormDB != nil {
		if sqlDB, err := gormDB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return nil
}

func refreshProviderToken(dbService services.DatabaseService, provider *models.Provider) error {
	// Check if refresh token is available
	if provider.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	if provider.TokenURL == "" || provider.ClientID == "" {
		return fmt.Errorf("missing required OAuth config (token URL or client ID)")
	}

	// Build OAuth2 config
	oauthCfg := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: provider.TokenURL},
		RedirectURL:  provider.RedirectURL,
		Scopes:       common.SplitScopes(provider.Scopes),
	}

	// Create token with refresh token
	token := &oauth2.Token{
		AccessToken:  provider.AccessToken,
		RefreshToken: provider.RefreshToken,
		Expiry:       timePtrValue(provider.TokenExpiry),
	}

	// Refresh the token
	newToken, err := oauthCfg.TokenSource(context.Background(), token).Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update provider with new token info
	provider.AccessToken = newToken.AccessToken
	provider.RefreshToken = newToken.RefreshToken
	provider.TokenExpiry = &newToken.Expiry

	// Save the updated token to database using DAO
	dao := dbService.GetDAO()
	if err := dao.Provider.Save(provider); err != nil {
		return fmt.Errorf("failed to save updated token: %w", err)
	}

	return nil
}

func timePtrValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
