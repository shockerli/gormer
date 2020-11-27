package gormer

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// ChunkByIDMaxMin process data in chunks, scope by id
func ChunkByIDMaxMin(size int, db *gorm.DB, dest interface{}, callback func(), l Logger) {
	if l == nil {
		l = new(DefaultLogger)
	}

	// query the maximum and minimum primary key id that satisfy the criteria
	var stat = struct {
		MaxID int `json:"max_id"`
		MinID int `json:"min_id"`
	}{}
	err := db.Debug().
		Select("MAX(id) AS max_id, MIN(id) AS min_id").
		Take(&stat).Error

	if (err != nil && gorm.IsRecordNotFoundError(err)) || stat.MaxID == 0 {
		l.Info("no matching data...")
		return
	}
	if err != nil {
		l.Error(fmt.Sprintf("query MaxId and MinId error: %v", err))
		return
	}

	l.Info(fmt.Sprintf("query result: MinId(%d), MaxId(%d)", stat.MinID, stat.MaxID))

	// store the max id of last loop
	var lastMaxID = stat.MinID
	var loop = 0

	for {
		loop++

		if lastMaxID > stat.MaxID {
			l.Info("data processing is completed...")
			break
		}

		// start at MinId, end at MaxId
		lt := lastMaxID + size
		if lt > stat.MaxID {
			lt = stat.MaxID + 1
		}

		// paging through id range coverage
		res := db.Where("id >= ?", lastMaxID).
			Where("id < ?", lt).
			Scan(dest)

		l.Info(fmt.Sprintf("loop: %d, query result %d <= id < %d, count: %d, err: %v", loop, lastMaxID, lt, res.RowsAffected, res.Error))

		lastMaxID += size

		// no data queried, continue next cycle
		// if the id is discontinuous, it may detect that the data is empty,
		//          but it does not mean that the loop is closed
		if res.Error != nil || res.RowsAffected <= 0 {
			continue
		}

		// custom process by callback
		callback()

	}

}
