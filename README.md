# gormer
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shockerli/gormer)](https://pkg.go.dev/github.com/shockerli/gormer)
> Useful scopes and tools for [GORM V1](https://github.com/jinzhu/gorm)


## Chunk
```go
db, err := gorm.Open(
    "mysql",
    "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s")
if err != nil {
    return
}

var data []YourObject
db = db.Table("your_object_table_name").Select("id, name").Where("bz_id = ?", 999)

gormer.ChunkByIDMaxMin(50, db, &data, func() {
    for _, item := range data {
        print(item.ID, ", ")
    }
    println("\n")
}, nil)
```

## Pager
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

## Order
```go
type QueryUserListParams struct {
    gormer.OrderParam
    Name string `json:"name"`
}
var params = QueryUserListParams{}

db = db.Table("your_object_table_name").Scopes(params.Order())
```

## sql-builder-gen
- Install

```shell
go install github.com/shockerli/gormer/sql-builder-gen
```

- Source code

> source code of `app/field/fields.go`

```go
//go:generate sql-builder-gen
package field

import "app/enum"

//go:sql-builder-gen -f=ALL
type IDX struct {
    ID int64 `json:"id" gorm:"column:id"`
}

//go:sql-builder-gen -f=EQ,IN
type DispatchTypeX struct {
    DispatchType enum.DispatchType `json:"dispatch_type" sql:"column:dispatch_type"`
}
```

### Generate code

> execute command: `go generate app/field/fields.go`
>
> generate codes to `app/field/fields_sql_builder_gen.go`:

```go
// Code generated by github.com/shockerli/gormer/sql-builder-gen DO NOT EDIT

// Package field SQL build helper functions for field
package field

import (
  "github.com/jinzhu/gorm"
  "github.com/shockerli/gormer/tmp/enum"
)

// ****************** ID ****************** //

// Eq WHERE id = ?
func (i IDX) Eq(v ...int64) func(db *gorm.DB) *gorm.DB {
    if len(v) == 0 {
        v = append(v, i.ID)
    }
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("id = ?", v[0])
    }
}

// Gt WHERE id > ?
func (i IDX) Gt(v ...int64) func(db *gorm.DB) *gorm.DB {
    if len(v) == 0 {
        v = append(v, i.ID)
    }
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("id > ?", v[0])
    }
}

// ****************** DispatchType ****************** //

// Eq WHERE dispatch_type = ?
func (i DispatchTypeX) Eq(v ...enum.DispatchType) func(db *gorm.DB) *gorm.DB {
    if len(v) == 0 {
        v = append(v, i.DispatchType)
    }
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("dispatch_type = ?", v[0])
    }
}
```

### Use code
> use `Scope()` build where condition

```go
type ObjectItem struct {
    field.IDX
    field.DispatchTypeX
}

func (ObjectItem) TableName() string {
    return "object_item_01"
}

func main() {
    m := ObjectItem{}
    err := DB().Model(m).Scopes(m.IDX.Eq(1146)).Take(&m).Error
    s, _ := json.Marshal(m)
    spew.Dump(err, m)
    println(string(s), "\n")

    json.Unmarshal([]byte(`{"id":999}`), &m)
    println(m.DispatchType)
}
```
