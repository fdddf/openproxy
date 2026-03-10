package services

import (
	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/gofiber/fiber/v2"
)

// StatsServiceImpl implements the StatsService interface
type StatsServiceImpl struct {
	dbService DatabaseService
}

// NewStatsService creates a new StatsService instance
func NewStatsService(dbService DatabaseService) StatsService {
	return &StatsServiceImpl{
		dbService: dbService,
	}
}

func (s *StatsServiceImpl) HandleGetStats(c *fiber.Ctx) error {
	dao := s.dbService.GetDAO()

	// Get counts for different entities
	var apiKeyCount int64
	var providerCount int64
	var requestCount int64
	var successRequestCount int64

	var err error

	if apiKeyCount, err = dao.APIKey.Count(); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to count API keys")
	}
	if providerCount, err = dao.Provider.Count(); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to count providers")
	}
	if requestCount, err = dao.Request.Count(); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to count requests")
	}

	if successRequestCount, err = dao.Request.Where(
		dao.Request.StatusCode.Gte(200),
		dao.Request.StatusCode.Lt(300),
	).Count(); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to count successful requests")
	}

	successRate := 0.0
	if requestCount > 0 {
		successRate = float64(successRequestCount) / float64(requestCount)
	}

	stats := dto.Stats{
		TotalRequests:   requestCount,
		ActiveAPIKeys:   apiKeyCount,
		ActiveProviders: providerCount,
		SuccessRate:     successRate,
	}

	return c.JSON(stats)
}
