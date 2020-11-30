package gormer

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Schema table schema info
type Schema struct {
	RawDDL           string   `json:"raw_ddl"`
	TableName        string   `json:"table_name"`
	Engine           string   `json:"engine"`
	AutoIncrement    int64    `json:"auto_increment"`
	DefaultCharset   string   `json:"default_charset"`
	DefaultCollation string   `json:"default_collation"`
	Comment          string   `json:"comment"`
	PrimaryKey       string   `json:"primary_key"`
	Fields           []*Field `json:"fields"`
	Keys             []Key    `json:"keys"`
}

// Key table index key
type Key struct {
	Name        string   `json:"name"`
	IndexType   string   `json:"index_type"`
	IndexMethod string   `json:"index_method"`
	Comment     string   `json:"comment"`
	Fields      []string `json:"fields"`
}

// Field table field info
type Field struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Length        int64       `json:"length"`
	Decimal       int64       `json:"decimal"`
	NotNull       bool        `json:"not_null"`
	Unsigned      bool        `json:"unsigned"`
	AutoIncrement int64       `json:"auto_increment"`
	Default       interface{} `json:"default"`
	Comment       string      `json:"comment"`
}

// Markdown return markdown format content
func (Schema) Markdown() string {
	return ""
}

// JSON return json format content
func (s Schema) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// NewDDL new ddl for table, remove auto_increment
func (s Schema) NewDDL() string {
	// 去除自增项
	reg, err := regexp.Compile(`(?i)\sAUTO_INCREMENT=\d*\s`)
	if err != nil {
		return s.RawDDL
	}

	res := reg.FindStringSubmatch(s.RawDDL)
	if len(res) >= 1 {
		return strings.Replace(s.RawDDL, res[0], " ", 1)
	}

	return s.RawDDL
}
