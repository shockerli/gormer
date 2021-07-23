package gormer

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// ChunkByIDMaxMin process data in chunks, scope by id
func ChunkByIDMaxMin(size int64, db *gorm.DB, dest interface{}, callback func(), l Logger) {
	if l == nil {
		l = new(DefaultLogger)
	}
	startTime := time.Now().UnixNano()

	// query the maximum and minimum primary key id that satisfy the criteria
	var stat = struct {
		MaxID sql.NullInt64 `json:"max_id"`
		MinID sql.NullInt64 `json:"min_id"`
	}{}
	err := db.NewScope(db).DB().Select("MAX(id) AS max_id, MIN(id) AS min_id").Take(&stat).Error

	if err != nil {
		l.Error(fmt.Sprintf("query MinId and MaxId error: %v", err))
		return
	}
	if !stat.MaxID.Valid || !stat.MinID.Valid {
		l.Info("no matching data...MinId(null), MaxId(null)")
		return
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
		res := db.NewScope(db).DB().
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
		callback()
	}

	usedTime := fmt.Sprintf("%.2fms", float64(time.Now().UnixNano()-startTime)/1e6)
	l.Info(fmt.Sprintf("data processing is completed...Used: %s, TotalCount: %d", usedTime, totalCount))

}
