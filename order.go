package gscope

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// 统一的排序参数
type OrderParam struct {
	OrderBy   string `json:"order_by"` // 可以是具体字段名，也可以是逗号分隔的多个排序
	OrderType string `json:"order_type"`
}

// 处理排序
func ScopeOrderBy(param *OrderParam) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if param != nil && param.OrderBy != "" {
			order := param.OrderBy
			if !strings.Contains(param.OrderBy, ",") && strings.ToLower(param.OrderType) == "desc" {
				order += " DESC"
			}
			return db.Order(order)
		}
		return db
	}
}
