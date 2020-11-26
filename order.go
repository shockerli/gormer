package gscope

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// OrderParam
type OrderParam struct {
	OrderBy   string `json:"order_by"` // 可以是具体字段名，也可以是逗号分隔的多个排序
	OrderType string `json:"order_type"`
}

// OrderBy scope: order by
func (p *OrderParam) Order() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p != nil && p.OrderBy != "" {
			order := p.OrderBy
			if strings.ToUpper(p.OrderType) == "DESC" &&
				!strings.Contains(p.OrderBy, ",") &&
				!strings.Contains(strings.ToUpper(p.OrderBy), " ASC") &&
				!strings.Contains(strings.ToUpper(p.OrderBy), " DESC") {

				order += " DESC"
			}
			return db.Order(order)
		}
		return db
	}
}
