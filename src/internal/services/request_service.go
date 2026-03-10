package services

import (
	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/gofiber/fiber/v2"
)

// RequestServiceImpl implements the RequestService interface
type RequestServiceImpl struct {
	dbService DatabaseService
}

// NewRequestService creates a new RequestService instance
func NewRequestService(dbService DatabaseService) RequestService {
	return &RequestServiceImpl{
		dbService: dbService,
	}
}

func (r *RequestServiceImpl) HandleGetRequests(c *fiber.Ctx) error {
	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize

	dao := r.dbService.GetDAO()
	requests, total, err := dao.Request.
		Order(dao.Request.CreatedAt.Desc()).
		FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get requests")
	}

	resp := make([]dto.Request, 0, len(requests))
	for _, req := range requests {
		resp = append(resp, dto.NewRequest(req))
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}
