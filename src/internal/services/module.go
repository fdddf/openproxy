package services

import (
	"go.uber.org/fx"
)

// Module provides all the services for the application
var Module = fx.Options(
	fx.Provide(
		NewConfigService,
		NewDatabaseService,
		NewProviderService,
		NewChatService,
		NewAPIService,
		NewModelService,
		NewHealthCheckService,
		// Custom provider for NewProviderConfigService that depends on HealthCheckService
		func(dbService DatabaseService, healthCheckService HealthCheckService) ProviderConfigService {
			return NewProviderConfigService(dbService, healthCheckService)
		},
		NewLogService,
		NewRequestService,
		NewStatsService,
		NewSettingsService,
		NewSecretService,
		NewRateLimitService,
	),
)
