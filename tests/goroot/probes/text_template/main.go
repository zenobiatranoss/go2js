package main

import (
	"fmt"
	"os"
	"text/template"
)

func main() {
	t := template.Must(template.New("x").Parse("Hello {{.Name}}, you have {{.Count}} messages."))
	err := t.Execute(os.Stdout, struct {
		Name  string
		Count int
	}{"ann", 7})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()

	ift := template.Must(template.New("ift").Parse("{{if .}}yes{{else}}no{{end}}"))
	for _, data := range []bool{true, false} {
		_ = ift.Execute(os.Stdout, data)
		fmt.Println()
	}

	rng := template.Must(template.New("rng").Parse("{{range .}}[{{.}}]{{end}}"))
	_ = rng.Execute(os.Stdout, []string{"a", "b"})
	fmt.Println()

	dot := template.Must(template.New("dot").Parse("{{.}}"))
	_ = dot.Execute(os.Stdout, 41)
	fmt.Println()
}
