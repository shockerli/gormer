package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"
)

// Field database field
type Field struct {
	Name   string   // base struct name
	Field  string   // field name
	Type   string   // field type
	Column string   // database column name
	Import string   // imported packages
	Fns    []string // need generate functions
}

const suffix = "_sql_builder_gen"

var file = os.Getenv("GOFILE")

func main() {
	fields := parse()

	codeFile := strings.TrimSuffix(file, ".go") + suffix + ".go"
	codes, err := genCode(fields)
	// println(string(codes))
	checkErr(err)
	checkErr(ioutil.WriteFile(codeFile, codes, 0644))
}

func parse() []Field {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	checkErr(err)

	imps := parseImports(f.Imports)

	var fields []Field
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			var fns []string
			var gen = false
			if decl.Doc == nil {
				continue
			}
			for _, comment := range decl.Doc.List {
				if strings.HasPrefix(comment.Text, "//go:sql-builder-gen") {
					gen = true

					// func
					fns = parseFunc(comment.Text)
				}
			}
			if !gen {
				continue
			}

			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					switch t := spec.Type.(type) {
					case *ast.StructType:
						if len(t.Fields.List) != 1 {
							continue
						}

						field := Field{
							// Name
							Name: spec.Name.String(),
						}
						for _, f := range t.Fields.List {
							// Field
							if len(f.Names) != 1 {
								continue
							}
							for _, name := range f.Names {
								field.Field = name.String()
							}

							// Type
							switch ft := f.Type.(type) {
							case *ast.Ident:
								field.Type = ft.String()
							case *ast.SelectorExpr:
								pkg := ft.X.(*ast.Ident).String()
								field.Type = pkg + "." + ft.Sel.String()

								// Import
								field.Import = imps[pkg]
							}

							// Column
							tags := parseTagSetting(reflect.StructTag(f.Tag.Value))
							if col, ok := tags["COLUMN"]; ok {
								field.Column = col
							}

							// func
							field.Fns = fns

							fields = append(fields, field)
						}
					}
				}
			}
		}
	}

	return fields
}

func genCode(fields []Field) ([]byte, error) {
	var (
		t   *template.Template
		err error
		buf = bytes.NewBufferString("")
	)

	// pkg & import
	t, err = template.New("").Parse(tplPkg)
	if err != nil {
		return nil, fmt.Errorf("template init err %w", err)
	}
	err = t.Execute(buf, map[string]interface{}{
		"pkg":    os.Getenv("GOPACKAGE"), // source package
		"fields": fields,
	})

	// func
	for _, f := range fields {
		buf.WriteString(fmt.Sprintf(`// ****************** %s ****************** //
`, f.Field))

		for _, fn := range f.Fns {
			if _, ok := tplFunc[fn]; !ok {
				continue
			}

			t, err = template.New("").Parse(tplFunc[fn])
			if err != nil {
				return nil, fmt.Errorf("template %s parse err %w", fn, err)
			}

			err = t.Execute(buf, f)
			if err != nil {
				return nil, fmt.Errorf("template %s execute err %w", fn, err)
			}
		}
	}

	return format.Source(buf.Bytes())
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("err: %+v", err))
	}
}

// copy from github.com/jinzhu/gorm/model_struct.go:parseTagSetting
func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get("gorm")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}

func parseImports(imps []*ast.ImportSpec) map[string]string {
	var m = make(map[string]string)
	for _, imp := range imps {
		var path = imp.Path.Value
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
			// build path with alias
			path = name + " " + path
		} else {
			// parse name from path
			if strings.Contains(path, "/") {
				name = strings.TrimRight(path[strings.LastIndex(path, "/")+1:], `"`)
			} else {
				name = path
			}
		}

		m[name] = path
	}

	return m
}

func parseFunc(comment string) []string {
	var cli = flag.NewFlagSet("x", flag.ExitOnError)
	var fn string
	cli.StringVar(&fn, "f", "EQ", "")
	err := cli.Parse(strings.Split(comment, " ")[1:])
	if err != nil {
		checkErr(fmt.Errorf("params pasrse err %w", err))
	}

	var fns []string
	var exist = make(map[string]struct{})
	for _, s := range strings.Split(fn, ",") {
		s = strings.ToUpper(strings.TrimSpace(s)) // trim and upper
		// empty
		if s == "" {
			continue
		}
		// unique
		if _, ok := exist[s]; ok {
			continue
		}

		// ALL=all func
		if s == ALL {
			fns = fns[0:0]
			for f := range tplFunc {
				fns = append(fns, f)
			}
			break
		}

		exist[s] = struct{}{}
		fns = append(fns, s)
	}
	sort.Strings(fns) // sorted by func name

	return fns
}
