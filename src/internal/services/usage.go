package services

import (
	"github.com/fdddf/openproxy/internal/dao"
	"gorm.io/gen/field"
)

// usageCounts returns how many recorded requests reference each of the given
// ids, keyed by id.
//
// The Models and API Keys views both render a "usage" column that was
// previously hardcoded to zero in the DTOs. One grouped query per page keeps
// that column honest without an N+1.
func usageCounts(q *dao.Query, column field.Uint, ids []uint) map[uint]int {
	counts := make(map[uint]int, len(ids))
	if q == nil || len(ids) == 0 {
		return counts
	}

	var rows []struct {
		ID    uint `gorm:"column:ref_id"`
		Total int  `gorm:"column:total"`
	}

	if err := q.Request.
		Select(column.As("ref_id"), q.Request.ID.Count().As("total")).
		Where(column.In(ids...)).
		Group(column).
		Scan(&rows); err != nil {
		// A failed count should not fail the listing; report zero usage.
		return counts
	}

	for _, row := range rows {
		counts[row.ID] = row.Total
	}
	return counts
}
