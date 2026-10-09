package integration_test

import "testing"

// text/template is walked the way go walks it: a missing map key is left as
// <no value> in an action but <nil> through print, a field of a struct that
// is not there is an error, a range over a map runs its keys in order, a
// define is a template you can call, and a parse or exec fault reads the same
// line and column go reads.
func TestTextTemplateParity(t *testing.T) {
	runParityTest(t, `package main

import (
	"bytes"
	"fmt"
	"text/template"
)

type Item struct {
	Name  string
	Price float64
	Tags  []string
}

func run(name, text string, data interface{}) {
	t, err := template.New(name).Parse(text)
	if err != nil {
		fmt.Printf("%s parse err=%v\n", name, err)
		return
	}
	var out bytes.Buffer
	err = t.Execute(&out, data)
	fmt.Printf("%s out=%q err=%v\n", name, out.String(), err)
}

func main() {
	item := Item{"tea", 2.5, []string{"a", "b"}}
	run("field", "{{.Name}}|{{.Price}}|{{len .Tags}}", item)
	run("chain", "{{.Tags}} {{index .Tags 1}}", item)
	run("mapmiss", "{{.k}} {{.missing}}", map[string]string{"k": "v"})
	run("with", "{{with .}}{{.Name}}{{end}}", item)
	run("withelse", "{{with .Missing}}x{{else}}none{{end}}", item)
	run("ifelse", "{{if .}}yes{{else}}no{{end}}", 0)
	run("elseif", "{{if eq .A 1}}one{{else if eq .A 2}}two{{else}}other{{end}}", map[string]int{"A": 2})
	run("root", "{{$x := 1}}{{$.Name}} {{$x}}", item)
	run("range", "{{range $i, $v := .}}{{$i}}={{$v}} {{end}}", []string{"a", "b"})
	run("rangemap", "{{range $k, $v := .}}{{$k}}={{$v}};{{end}}", map[string]int{"b": 2, "a": 1})
	run("rangeint", "{{range 3}}{{.}}{{end}}", nil)
	run("pipelines", "{{.Name | printf \"<%s>\" | printf \"[%s]\"}}", item)
	run("trim", "a{{- .Name -}}b", item)
	run("nested", "{{define \"inner\"}}<{{.}}>{{end}}{{template \"inner\" .}}", "q")
	run("tplarg", "{{define \"t\"}}{{$}}:{{.}}{{end}}{{template \"t\" 9}}", 3)
	run("tplk", "{{define \"t\"}}{{$}}:{{.}}{{end}}{{template \"t\"}}", 3)
	run("block", "a{{block \"b\" .}}<{{.}}>{{end}}b", "z")
	run("and", "{{and .A .B}}", struct{ A, B int }{0, 5})
	run("or", "{{or .A .B}}", struct{ A, B int }{3, 0})
	run("printnil", "{{print .}}", nil)
	run("dotnil", "{{.}}", nil)
	run("index", "{{index .Tags 1}}", item)
	run("slice", "{{slice .Tags 0 1}}", item)
	run("html", "{{. | html}}", "<b>&'\"")
	run("js", "{{. | js}}", "a'b\"")
	run("url", "{{. | urlquery}}", "a b&c")
	run("fielderr", "{{with .Missing}}x{{else}}none{{end}}", item)
	run("arity", "{{len 1 2}}", nil)
	run("nilcmd", "{{nil}}", nil)
	run("rniter", "{{range 2.5}}x{{end}}", nil)
	run("undef", "{{template \"nope\"}}", nil)
	run("neit", "{{define \"t\"}}ab{{.Nope}}{{end}}{{template \"t\" 5}}", nil)
}
`)
}
