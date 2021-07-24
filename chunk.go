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
func ChunkByIDMaxMin(size int64, db *gorm.DB, dest interface{}, callback ChunkCallback, l Logger) (err error) {
	if l == nil {
		l = new(DefaultLogger)
	}
	startTime := time.Now().UnixNano()
	tableName := TableName(db)

	// query the maximum and minimum primary key id that satisfy the criteria
	var stat = struct {
		MaxID sql.NullInt64 `json:"max_id"`
		MinID sql.NullInt64 `json:"min_id"`
	}{}
	err = db.NewScope(db.Value).DB().Table(tableName).
		Select("MAX(id) AS max_id, MIN(id) AS min_id").
		Take(&stat).Error

	if err != nil {
		l.Error(fmt.Sprintf("query MinId and MaxId error: %v", err))
		return
	}
	if !stat.MaxID.Valid || !stat.MinID.Valid {
		l.Info("no matching data...MinId(null), MaxId(null)")
		return nil
	}

	l.Info(fmt.Sprintf("query result: MinId(%d), MaxId(%d)", stat.MinID.Int64, stat.MaxID.Int64))

	// store the max id of last loop
	var lastMaxID = stat.MinID.Int64
	var loop = 0
	var totalCount int64

	for {
		loop++

		if lastMaxID > stat.MaxID.Int64 {
			break
		}

		// start at MinId, end at MaxId
		lt := lastMaxID + size
		if lt > stat.MaxID.Int64 {
			lt = stat.MaxID.Int64 + 1
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
