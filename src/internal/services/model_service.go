package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ModelServiceImpl implements the ModelService interface
type ModelServiceImpl struct {
	dbService DatabaseService
}

// NewModelService creates a new ModelService instance
func NewModelService(dbService DatabaseService) ModelService {
	return &ModelServiceImpl{
		dbService: dbService,
	}
}

func (m *ModelServiceImpl) HandleGetModels(c *fiber.Ctx) error {
	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize
	providerID := c.QueryInt("providerId", 0)

	dao := m.dbService.GetDAO()
	query := dao.Model.Order(dao.Model.CreatedAt.Desc())
	if providerID > 0 {
		query = query.Where(dao.Model.ProviderID.Eq(uint(providerID)))
	}

	modelsList, total, err := query.FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get models")
	}

	ids := make([]uint, 0, len(modelsList))
	for _, model := range modelsList {
		ids = append(ids, model.ID)
	}
	counts := usageCounts(dao, dao.Request.ModelID, ids)

	resp := make([]dto.Model, 0, len(modelsList))
	for _, model := range modelsList {
		item := dto.NewModel(model)
		item.UsageCount = counts[model.ID]
		resp = append(resp, item)
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}

func (m *ModelServiceImpl) HandleCreateModel(c *fiber.Ctx) error {
	var req struct {
		ProviderID uint   `json:"providerId"`
		Name       string `json:"name"`
		MappedName string `json:"mappedName"`
		IsActive   *bool  `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := m.validateModelPayload(c, req.ProviderID, req.Name, req.MappedName, 0); err != nil {
		return err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	model := models.Model{
		ProviderID: req.ProviderID,
		Name:       req.MappedName,
		RealModel:  req.Name,
		IsActive:   isActive,
	}

	if err := m.dbService.GetDAO().Model.Create(&model); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create model")
	}

	return c.JSON(dto.NewModel(&model))
}

func (m *ModelServiceImpl) HandleUpdateModel(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid model id")
	}

	dao := m.dbService.GetDAO()
	model, err := dao.Model.Where(dao.Model.ID.Eq(uint(id))).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "Model not found")
	}

	var req struct {
		ProviderID uint   `json:"providerId"`
		Name       string `json:"name"`
		MappedName string `json:"mappedName"`
		IsActive   *bool  `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := m.validateModelPayload(c, req.ProviderID, req.Name, req.MappedName, model.ID); err != nil {
		return err
	}

	model.ProviderID = req.ProviderID
	model.RealModel = req.Name
	model.Name = req.MappedName
	if req.IsActive != nil {
		model.IsActive = *req.IsActive
	}

	if err := dao.Model.Save(model); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to update model")
	}

	return c.JSON(dto.NewModel(model))
}

func (m *ModelServiceImpl) HandleDeleteModel(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid model id")
	}

	dao := m.dbService.GetDAO()
	if _, err := dao.Model.Where(dao.Model.ID.Eq(uint(id))).Delete(&models.Model{}); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to delete model")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (m *ModelServiceImpl) validateModelPayload(c *fiber.Ctx, providerID uint, realName, mappedName string, excludeID uint) error {
	realName = strings.TrimSpace(realName)
	mappedName = strings.TrimSpace(mappedName)

	if providerID == 0 || realName == "" || mappedName == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Provider, original name, and mapped name are required")
	}

	dao := m.dbService.GetDAO()
	if _, err := dao.Provider.Where(dao.Provider.ID.Eq(providerID)).First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusBadRequest, "Provider not found")
		}
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load provider")
	}

	// Use DAO to check for existing model
	query := dao.Model.Where(dao.Model.ProviderID.Eq(providerID), dao.Model.Name.Eq(mappedName))
	if excludeID != 0 {
		query = query.Where(dao.Model.ID.Neq(excludeID))
	}

	_, err := query.First()
	if err == nil {
		return common.HandleError(c, fiber.StatusBadRequest, "A mapping with that name already exists for this provider")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to validate model mapping")
	}

	return nil
}
