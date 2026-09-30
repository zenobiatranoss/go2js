package integration_test

import "testing"

// A block action is decided by the pipeline that follows its name, not by the
// body that comes after it. Reading the body instead leaves the action empty, so
// an if always takes its first branch and a range never enters its body at all.
func TestTemplateBlockActionsReadTheirOwnPipeline(t *testing.T) {
	runParityTest(t, `package main

import (
	"os"
	"text/template"
)

type Item struct {
	Name  string
	Price int
	Tags  []string
}

func render(name, text string, data any) {
	tpl, err := template.New(name).Parse(text)

	if err != nil {
		os.Stdout.WriteString("parse: " + err.Error() + "\n")
		return
	}

	if err := tpl.Execute(os.Stdout, data); err != nil {
		os.Stdout.WriteString("execute: " + err.Error() + "\n")
	}
}

func main() {
	item := Item{Name: "x", Price: 9, Tags: []string{"a", "b"}}

	render("one", "Hello {{.Name}}!\n", item)
	render("two", "{{range .Tags}}[{{.}}]{{end}}|{{len .Tags}}\n", item)
	render("three", "{{if .Tags}}yes{{else}}no{{end}}|{{if not .Tags}}no{{else}}yes{{end}}\n", item)
	render("four", "{{range .Tags}}{{if eq . \"a\"}}A{{else if eq . \"b\"}}B{{end}}{{end}}\n", item)

	empty := Item{}
	render("five", "{{if .Tags}}yes{{else}}no{{end}}|{{len .Tags}}|{{range .Tags}}x{{end}}\n", empty)

	// A number reaches the output as arithmetic left it, and is written the way
	// the language writes a number rather than the way JavaScript does.
	render("six", "{{len .Tags}} {{.Price}} {{.Price}}0\n", item)

	// A block walks the values it is given and reports the position, and a
	// variable stays in scope for the body that follows it.
	render("seven", "{{range $i, $t := .Tags}}{{$i}}:{{$t}} {{end}}|{{$.Name}}\n", item)

	// A with that finds a value switches the dot to it for the length of its body.
	render("eight", "{{with .Tags}}{{len .}}{{end}}|{{with .Name}}{{.}}{{end}}\n", item)
}
`)
}
