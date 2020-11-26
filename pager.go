package gormer

import "github.com/jinzhu/gorm"

// PageParam uniform paging parameters
type PageParam struct {
	CurrPage   int `json:"curr_page" binding:"number"` // (optional) current page number, default 1
	PageSize   int `json:"page_size" binding:"number"` // (optional) number of per page, default 10
	IgnorePage int `json:"ignore_page,omitempty"`      // (optional) no paging and no total count, while not 0
}

// Init init param
func (p *PageParam) Init() {
	if p == nil {
		p = &PageParam{
			PageSize: 10,
			CurrPage: 1,
		}
	}

	if p.CurrPage <= 0 {
		p.CurrPage = 1
	}

	if p.PageSize <= 0 {
		p.PageSize = 10
	}
}

// Page pager scope
func (p *PageParam) Page() func(db *gorm.DB) *gorm.DB {
	// ignore page and count
	if p.IgnorePage != 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	p.Init()
	offset := (p.CurrPage - 1) * p.PageSize

	return func(db *gorm.DB) *gorm.DB {
		if offset > 0 {
			db = db.Offset(offset)
		}
		if p.PageSize > 0 {
			db = db.Limit(p.PageSize)
		}
		return db
	}
}

// PageResult unified paging response structure
type PageResult struct {
	PageParam
	TotalCount int `json:"total_count"`
	CurrCount  int `json:"curr_count"`
}

// Count count result
func (p *PageResult) Count(db *gorm.DB) error {
	if p.IgnorePage != 0 {
		return nil
	}

	return db.Count(&p.TotalCount).Error
}

// Scan scan result and count
func (p *PageResult) Scan(db *gorm.DB, dest interface{}) (err error) {
	db = db.Scan(dest)
	err = db.Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	if err != nil {
		return
	}

	p.CurrCount = int(db.RowsAffected)
	if p.IgnorePage != 0 {
		p.TotalCount = p.CurrCount
	} else {
		return p.Count(db)
	}

	return
}
