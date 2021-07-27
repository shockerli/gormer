package gormer

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

var regexpUnsigned, _ = regexp.Compile(`(?i)\sUNSIGNED`)
var regexpCharset, _ = regexp.Compile(`(?i)\sCHARACTER SET\s(\w+)`)
var regexpCollate, _ = regexp.Compile(`(?i)\sCOLLATE\s(\w+)`)

// Schema table schema info
type Schema struct {
	RawDDL           string    `json:"-"`
	TableName        string    `json:"table_name"`
	Engine           string    `json:"engine"`
	AutoIncrement    uint64    `json:"auto_increment"`
	DefaultCharset   string    `json:"default_charset"`
	DefaultCollation string    `json:"default_collation"`
	Comment          string    `json:"comment"`
	Columns          []*Column `json:"columns"`
	Keys             []*Key    `json:"keys"`
}

// Key table index key
type Key struct {
	Name        string   `json:"name"`
	IndexType   string   `json:"index_type"`
	IndexMethod string   `json:"index_method"`
	Comment     string   `json:"comment"`
	Columns     []string `json:"columns"`
}

// Column table field info
type Column struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	NotNull       bool        `json:"not_null"`
	Unsigned      bool        `json:"unsigned"`
	AutoIncrement bool        `json:"auto_increment"`
	Charset       string      `json:"charset"`
	Collate       string      `json:"collate"`
	Default       interface{} `json:"default"`
	Comment       string      `json:"comment"`
}

// Parse parse schema info from raw DDL
func (s *Schema) Parse() error {
	stmt, err := sqlparser.Parse(s.RawDDL)
	if err != nil {
		return err
	}

	ddl, ok := stmt.(*sqlparser.CreateTable)
	if !ok || ddl.Action != "create" {
		return errors.New("DDL is not a CREATE statement")
	}

	// table name
	s.TableName = ddl.NewName.Name.String()

	// keys
	for _, v := range ddl.Constraints {
		var key = &Key{
			Name:      v.Name,
			IndexType: v.Type.String(),
			Columns:   nil,
		}
		for _, k := range v.Keys {
			key.Columns = append(key.Columns, k.String())
		}

		s.Keys = append(s.Keys, key)
	}

	// options
	for _, v := range ddl.Options {
		switch v.Type {
		case sqlparser.TableOptionEngine:
			s.Engine = v.StrValue
		case sqlparser.TableOptionCharset:
			s.DefaultCharset = v.StrValue
		case sqlparser.TableOptionAutoIncrement:
			s.AutoIncrement = v.UintValue
		case sqlparser.TableOptionCollate:
			s.DefaultCollation = v.StrValue
		case sqlparser.TableOptionComment:
			s.Comment = v.StrValue
		}
	}

	// columns
	for _, v := range ddl.Columns {
		var col = &Column{
			Name: v.Name,
			Type: v.Type,
		}
		unsigned := regexpUnsigned.FindStringSubmatch(col.Type)
		if len(unsigned) >= 1 {
			col.Unsigned = true
			col.Type = strings.Replace(col.Type, unsigned[0], "", 1)
		}
		charset := regexpCharset.FindStringSubmatch(col.Type)
		if len(charset) >= 2 {
			col.Charset = charset[1]
			col.Type = strings.Replace(col.Type, charset[0], "", 1)
		}
		collate := regexpCollate.FindStringSubmatch(col.Type)
		if len(collate) >= 2 {
			col.Collate = collate[1]
			col.Type = strings.Replace(col.Type, collate[0], "", 1)
		}

		// column option
		for _, opt := range v.Options {
			switch opt.Type {
			case sqlparser.ColumnOptionDefaultValue:
				col.Default = strings.Trim(opt.Value, `"`)
			case sqlparser.ColumnOptionNotNull:
				col.NotNull = true
			case sqlparser.ColumnOptionNull:
				col.NotNull = false
			case sqlparser.ColumnOptionAutoIncrement:
				col.AutoIncrement = true
			case sqlparser.ColumnOptionComment:
				col.Comment = strings.Trim(opt.Value, `"`)
			}
		}

		s.Columns = append(s.Columns, col)
	}

	return nil
}

// Markdown return markdown format content
func (s Schema) Markdown() string {
	var ms = `
- Table:    %s
- Comment:  %s
- Charset:  %s
- Engine:   %s

| Column | Type | Not Null | Default | Comment |
| --- | --- | :---: | --- | --- |
`

	ms = fmt.Sprintf(ms, s.TableName, s.Comment, s.DefaultCharset, s.Engine)

	for _, f := range s.Columns {
		var notNull = "❎"
		if f.NotNull {
			notNull = "✅"
		}
		var defaultVal = "-"
		if f.Default != nil {
			defaultVal = fmt.Sprintf(`"%v"`, f.Default)
		}

		ms += fmt.Sprintf(
			"| %s | %s | %s | %s | %s |\n",
			f.Name,
			f.Type,
			notNull,
			defaultVal,
			f.Comment,
		)
	}

	return ms
}

// JSON return json format content
func (s Schema) JSON() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// NewDDL new ddl for table, remove auto_increment
func (s Schema) NewDDL() string {
	// remove AUTO_INCREMENT
	reg, err := regexp.Compile(`(?i)\s*AUTO_INCREMENT=\d*\s*`)
	if err != nil {
		return s.RawDDL
	}

	res := reg.FindStringSubmatch(s.RawDDL)
	if len(res) >= 1 {
		return strings.TrimSpace(strings.Replace(s.RawDDL, res[0], " ", 1))
	}

	return s.RawDDL
}
