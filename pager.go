package gscope

import "github.com/jinzhu/gorm"

// 统一的分页参数
type PageParam struct {
	CurrPage   int `json:"curr_page" binding:"number"`
	PageSize   int `json:"page_size" binding:"number"`
	IgnorePage int `json:"ignore_page,omitempty"` // 不分页
}

func InitParam(param *PageParam) *PageParam {
	if param == nil {
		param = &PageParam{
			PageSize: 10,
			CurrPage: 1,
		}
	}

	if param.CurrPage <= 0 {
		param.CurrPage = 1
	}

	if param.PageSize <= 0 {
		param.PageSize = 10
	}

	return param
}

// 统一的分页响应结构
type PageResult struct {
	PageParam
	TotalCount int `json:"total_count"`
	CurrCount  int `json:"curr_count"`
}

// ScopePage 处理分页
func ScopePage(param *PageParam) func(db *gorm.DB) *gorm.DB {
	// 忽略分页&总数
	if param.IgnorePage != 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	param = InitParam(param)
	offset := (param.CurrPage - 1) * param.PageSize

	return func(db *gorm.DB) *gorm.DB {
		if offset > 0 {
			db = db.Offset(offset)
		}
		if param.PageSize > 0 {
			db = db.Limit(param.PageSize)
		}
		return db
	}
}
