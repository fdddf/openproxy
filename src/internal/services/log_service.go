package services

import (
	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/gofiber/fiber/v2"
)

// LogServiceImpl implements the LogService interface
type LogServiceImpl struct {
	dbService DatabaseService
}

// NewLogService creates a new LogService instance
func NewLogService(dbService DatabaseService) LogService {
	return &LogServiceImpl{
		dbService: dbService,
	}
}

func (l *LogServiceImpl) HandleGetLogs(c *fiber.Ctx) error {
	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize

	dao := l.dbService.GetDAO()
	logs, total, err := dao.Log.
		Order(dao.Log.CreatedAt.Desc()).
		FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get logs")
	}

	resp := make([]dto.Log, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, dto.NewLog(log))
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}
