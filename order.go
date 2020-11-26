package gormer

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// OrderParam
type OrderParam struct {
	OrderBy   string `json:"order_by"`   // (optional) column name, or SQL ORDER statement
	OrderType string `json:"order_type"` // (optional) ASC/DESC
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
