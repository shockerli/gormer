package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type Const struct {
	Name  string `json:"name"`
	Brief string `json:"brief"`
}

const suffix = "_gen"

var file = os.Getenv("GOFILE")

func main() {
	codes, err := genCode(parse())
	checkErr(err)

	codeFile := strings.TrimSuffix(file, ".go") + suffix + ".go"
	checkErr(ioutil.WriteFile(codeFile, codes, 0644))
}

func parse() map[string][]Const {
	var comments = make(map[string][]Const)

	// parse source code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	checkErr(err)

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments
	// and AST nodes.
	commentMap := ast.NewCommentMap(fset, f, f.Comments)
	for node := range commentMap {
		// only support: one line one const declare
		if spec, ok := node.(*ast.ValueSpec); !ok || len(spec.Names) != 1 {
			continue
		}
		spec := node.(*ast.ValueSpec)

		// only const declare
		ident := spec.Names[0]
		if ident.Obj.Kind != ast.Con {
			continue
		}

		// only custom type
		if spec.Type == nil {
			continue
		}
		t := spec.Type.(*ast.Ident).Name // custom type name

		// parse comment
		comments[t] = append(comments[t], Const{
			Name:  ident.Name,
			Brief: getComment(ident.Name, spec),
		})
	}

	return comments
}

func getComment(name string, vs *ast.ValueSpec) string {
	var buf bytes.Buffer

	var comments []*ast.Comment
	// line comment
	if vs.Comment != nil {
		comments = append(comments, vs.Comment.List...)
	}
	// block comment
	if vs.Doc != nil {
		comments = append(comments, vs.Doc.List...)
	}
	for _, comment := range comments {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)
		text = strings.TrimPrefix(text, name) // trim ident
		text = strings.TrimSpace(text)

		buf.WriteString(text)
		break // only first comment line
	}

	// replace any invisible with blanks
	bs := buf.Bytes()
	for i, b := range bs {
		switch b {
		case '\t', '\n', '\r':
			bs[i] = ' '
		}
	}
	return string(bs)
}

func genCode(types map[string][]Const) ([]byte, error) {
	var (
		t   *template.Template
		err error
		buf = bytes.NewBufferString("")
	)

	data := map[string]interface{}{
		"pkg":   os.Getenv("GOPACKAGE"),
		"types": types,
	}

	t, err = template.New("").Funcs(template.FuncMap{
		"firstLower": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
	}).Parse(tpl)
	if err != nil {
		return nil, fmt.Errorf("template parse err %w", err)
	}

	err = t.Execute(buf, data)
	if err != nil {
		return nil, fmt.Errorf("template execute err %w", err)
	}

	return format.Source(buf.Bytes())
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("err: %+v", err))
	}
}
