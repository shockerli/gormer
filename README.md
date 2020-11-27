# gormer
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shockerli/gormer)](https://pkg.go.dev/github.com/shockerli/gormer)
> Useful scopes and tools for [GORM v1](https://github.com/jinzhu/gorm)

## Examples

### Chunk
```go
db, err := gorm.Open(
    "mysql",
    "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s")
if err != nil {
    return
}

var data []YourObject
db = db.Table("your_object_table_name").Where("bz_id = ?", 999)

gormer.ChunkByIDMaxMin(50, db, &data, func() {
    for _, item := range data {
        print(item.ID, ", ")
    }
    println("\n")
}, nil)
```

### Pager
```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type QueryUserListParams struct {
    gormer.PageParam
    Name string `json:"name"`
}

type QueryUserListResult struct {
	gormer.PageResult
	List []User
}

var params = QueryUserListParams{}
var data = QueryUserListResult{}
data.PageParam = params.PageParam

db = db.Table("your_object_table_name").Where("bz_id = ?", 999)

data.Scan(db, &data.List)
```

### Order
```go
type QueryUserListParams struct {
    gormer.OrderParam
    Name string `json:"name"`
}
var params = QueryUserListParams{}

db = db.Table("your_object_table_name").Scopes(params.Order())
```
