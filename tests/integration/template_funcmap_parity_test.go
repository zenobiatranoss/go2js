package integration_test

import "testing"

// A function a program hands to a template through a FuncMap is found by name
// and called with what the template was written to hand it.
func TestTemplateFuncMapCustomFunctions(t *testing.T) {
	runParityTest(t, `package main

import (
	"os"
	"strings"
	"text/template"
)

func main() {
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
		"add":   func(a, b int) int { return a + b },
		"join":  strings.Join,
		"pair":  func(a, b string) string { return a + "/" + b },
	}

	data := map[string]interface{}{
		"Name":  "ann",
		"A":     2,
		"B":     3,
		"Items": []string{"x", "y", "z"},
	}

	t := template.Must(template.New("t").Funcs(funcs).Parse(
		"{{upper .Name}} {{add .A .B}} {{join .Items \", \"}} {{pair .Name \"!\"}}"))
	t.Execute(os.Stdout, data)
	os.Stdout.WriteString("\n")

	piped := template.Must(template.New("p").Funcs(funcs).Parse("{{.Name | upper}}"))
	piped.Execute(os.Stdout, data)
	os.Stdout.WriteString("\n")
}
`)
}
