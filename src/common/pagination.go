package common

import "github.com/gofiber/fiber/v2"

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// PaginatedResponse represents a standard paginated payload.
type PaginatedResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// NewPaginatedResponse builds a PaginatedResponse with sane defaults.
func NewPaginatedResponse[T any](items []T, total int64, page, pageSize int) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// GetPaginationParams extracts and normalizes pagination query params.
func GetPaginationParams(c *fiber.Ctx) (int, int) {
	page := c.QueryInt("page", defaultPage)
	if page < 1 {
		page = defaultPage
	}

	pageSize := c.QueryInt("pageSize", defaultPageSize)
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}
