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
	Extend []string // extend functions
}

const suffix = "_gen"

const name = "sql-gen"

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

	// Imports
	imps := parseImports(f.Imports)

	var fields []Field
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			var fns, ets []string
			var gen = false
			if decl.Doc == nil {
				continue
			}
			for _, comment := range decl.Doc.List {
				if strings.HasPrefix(comment.Text, "//go:"+name) {
					gen = true

					// func
					fn, ext := parseParams(comment.Text)
					fns = parseFunc(fn)
					ets = parseFunc(ext)
				}
			}
			if !gen {
				continue
			}

			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					switch t := spec.Type.(type) {
					case *ast.StructType: // only type struct
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

							// Func
							field.Fns = fns
							field.Extend = ets

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

	imps, nfs := clearData(fields)

	// pkg & import
	t, err = template.New("").Parse(tplPkg)
	if err != nil {
		return nil, fmt.Errorf("template init err %w", err)
	}
	err = t.Execute(buf, map[string]interface{}{
		"pkg":    os.Getenv("GOPACKAGE"), // source package
		"import": imps,
	})

	// func
	for _, f := range nfs {
		buf.WriteString(fmt.Sprintf(`// ****************** %s ****************** //
`, f.Field))

		// basic func
		for _, fn := range f.Fns {
			err = parseTpl(buf, fn, tplFunc[fn], f)
			if err != nil {
				return nil, err
			}
		}

		// extend func
		for _, ext := range f.Extend {
			err = parseTpl(buf, ext, tplExtend[ext], f)
			if err != nil {
				return nil, err
			}
		}
	}

	return format.Source(buf.Bytes())
}

func clearData(fs []Field) (imps []string, nfs []Field) {
	imps = append(imps, `"github.com/jinzhu/gorm"`)

	for _, f := range fs {
		// new Field
		var nf = Field{
			Name:   f.Name,
			Field:  f.Field,
			Type:   f.Type,
			Column: f.Column,
			Import: f.Import,
		}
		// import
		if f.Import != "" {
			imps = append(imps, f.Import)
		}

		// basic func
		for _, fn := range f.Fns {
			if _, ok := tplFunc[fn]; !ok {
				continue
			}
			nf.Fns = append(nf.Fns, fn)
		}

		// extend func
		for _, ext := range f.Extend {
			if _, ok := tplExtend[ext]; !ok {
				continue
			}

			switch true {
			case f.Type == "int64" && ext == TIME: // only support int64
				imps = append(imps, `"time"`, `"strconv"`) // need std package
			default:
				continue
			}
			nf.Extend = append(nf.Extend, ext)
		}

		nfs = append(nfs, nf)
	}
	imps = uniqueAndSortStrings(imps)

	return
}

func parseTpl(buf *bytes.Buffer, name, tpl string, data interface{}) error {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return fmt.Errorf("template %s parse err %w", name, err)
	}

	err = t.Execute(buf, data)
	if err != nil {
		return fmt.Errorf("template %s execute err %w", name, err)
	}
	return nil
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

func parseParams(comment string) (fn, ext string) {
	var cli = flag.NewFlagSet(name, flag.ExitOnError)
	cli.StringVar(&fn, "f", "EQ", "-f=EQ,IN")
	cli.StringVar(&ext, "t", "", "-t=TIME")
	err := cli.Parse(strings.Split(comment, " ")[1:])
	if err != nil {
		checkErr(fmt.Errorf("params parse err %w", err))
	}
	return
}

func parseFunc(fn string) []string {
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

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("err: %+v", err))
	}
}

func uniqueAndSortStrings(os []string) (ns []string) {
	var exist = make(map[string]struct{})
	for _, s := range os {
		if _, ok := exist[s]; ok {
			continue
		}
		ns = append(ns, s)
	}
	sort.Strings(ns)
	return
}
