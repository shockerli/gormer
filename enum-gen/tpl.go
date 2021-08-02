package main

var tpl = `
// Code generated by github.com/shockerli/gormer/enum-gen; DO NOT EDIT

// Package {{.pkg}} enum helpers
package {{.pkg}}

{{- range $type, $cons := .types}}
// ************ {{$type}} ************ //

var {{firstLower $type}}Map = map[{{$type}}]string{
{{- range $con := $cons}}
	{{$con.Name}}:"{{$con.Brief}}",
{{- end}}
}

// Check whether the type is exist
func (t {{$type}}) Check() bool {
	_, ok := {{$type}}Map()[t]
	return ok
}

// Desc return the desc of type
func (t {{$type}}) Desc() string {
	return {{$type}}Map()[t]
}

// {{$type}}Map return all of the types
func {{$type}}Map() map[{{$type}}]string {
	return {{firstLower $type}}Map
}
{{- end}}
`
