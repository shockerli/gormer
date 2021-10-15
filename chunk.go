package gormer

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// ChunkCallback chunk callback type
type ChunkCallback func(loop int) error

// ErrBreakChunk break the chunk while callback return error
var ErrBreakChunk = errors.New("break the chunk while")

// ChunkByIDMaxMin process data in chunks, scope by id
func ChunkByIDMaxMin(size int64, db *gorm.DB, dest interface{}, callback ChunkCallback, l Logger, extra ...func(db *gorm.DB) *gorm.DB) (err error) {
	if l == nil {
		l = new(DefaultLogger)
	}
	startTime := time.Now().UnixNano()
	tableName := TableName(db)

	maxID, minID, err := MaxMinID(db.Scopes(extra...))
	l.Info(fmt.Sprintf("query result: MinId(%d), MaxId(%d), ERR(%v)", minID, maxID, err))
	if err != nil {
		// ignore record not found
		if gorm.IsRecordNotFoundError(err) {
			err = nil
		}
		return
	}

	// store the max id of last loop
	var lastMaxID = minID
	var loop = 0
	var totalCount int64

	for {
		loop++

		if lastMaxID > maxID {
			break
		}

		// start at MinId, end at MaxId
		lt := lastMaxID + size
		if lt > maxID {
			lt = maxID + 1
		}

		// paging through id range coverage
		res := db.NewScope(db.Value).DB().Table(tableName).
			Where("? <= id AND id < ?", lastMaxID, lt).
			Scan(dest)

		l.Info(fmt.Sprintf("No.%d, query result %d <= id < %d, count: %d, err: %v", loop, lastMaxID, lt, res.RowsAffected, res.Error))

		lastMaxID += size
		totalCount += res.RowsAffected

		// no data queried, continue next cycle
		// if the id is discontinuous, it may detect that the data is empty,
		// but it does not mean that the loop is closed
		if res.Error != nil || res.RowsAffected <= 0 {
			continue
		}

		// custom process by callback
		// if callback return error wrap with ErrBreakChunk, break the while
		err = callback(loop)
		if err != nil {
			l.Error(fmt.Sprintf("No.%d, callback return ---> %v", loop, err))
			if errors.Is(err, ErrBreakChunk) {
				break
			}
		}
	}

	usedTime := fmt.Sprintf("%.2fms", float64(time.Now().UnixNano()-startTime)/1e6)
	l.Info(fmt.Sprintf("data processing is completed...Used: %s, TotalCount: %d", usedTime, totalCount))
	return
}

// TableName fetch table name from scope
func TableName(db *gorm.DB) string {
	if ts, ok := db.Value.(string); ok {
		return ts
	}
	return db.NewScope(db.Value).TableName()
}

// MaxMinID fetch the max and min ID for scope, support GROUP BY
func MaxMinID(db *gorm.DB) (max, min int64, err error) {
	tableName := TableName(db)

	// query the maximum and minimum primary key id that satisfy the criteria
	type Row struct {
		MaxID sql.NullInt64 `json:"max_id"`
		MinID sql.NullInt64 `json:"min_id"`
	}
	var stats []Row
	err = db.NewScope(db.Value).DB().Table(tableName). // new scope
								Select("MAX(id) AS max_id, MIN(id) AS min_id").
								Scan(&stats).Error // scan data to list, support GROUP BY

	// no records
	if err == nil && len(stats) == 0 {
		err = gorm.ErrRecordNotFound
	}

	// compare to the max and min
	for idx, v := range stats {
		if idx == 0 || v.MaxID.Int64 > max {
			max = v.MaxID.Int64
		}
		if idx == 0 || v.MinID.Int64 < min {
			min = v.MinID.Int64
		}
	}

	return
}
