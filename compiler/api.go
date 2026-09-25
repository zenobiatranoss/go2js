package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

type Compiler struct {
	Options Options
}

func New(options Options) (*Compiler, error) {
	options = options.Normalize()
	if err := options.Validate(); err != nil {
		return nil, err
	}
	return &Compiler{Options: options}, nil
}

func NewDefault() *Compiler {
	return &Compiler{Options: DefaultOptions()}
}

func (c *Compiler) CompileFile(path string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("empty source path")
	}

	parsed, err := ParseFile(path)
	if err != nil {
		return "", err
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		return "", err
	}

	output, err := javascript.EmitWithContextOptions(
		parsed.File,
		analysis.Types,
		analysis.Semantic,
		c.Options.Runtime,
	)
	if err != nil {
		return "", err
	}

	return c.wrap(output), nil
}

func (c *Compiler) CompilePackage(pkg *Package) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}
	if pkg == nil {
		return "", fmt.Errorf("nil package")
	}

	analysis, err := AnalyzePackage(pkg)
	if err != nil {
		return "", err
	}

	var parts []string
	for _, parsed := range pkg.Files {
		if parsed == nil || parsed.File == nil {
			return "", fmt.Errorf("package contains invalid file")
		}

		code, err := javascript.EmitWithContextOptions(
			parsed.File,
			analysis.Types,
			analysis.Semantic,
			c.Options.Runtime,
		)
		if err != nil {
			return "", err
		}

		parts = append(parts, code)
	}

	return c.wrap(strings.Join(parts, "\n")), nil
}

func (c *Compiler) CompileDirectory(dir string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("empty source directory")
	}

	if _, _, err := findModuleRoot(dir); err == nil {
		output, err := compileProjectWithOptions(dir, c.Options)
		if err != nil {
			return "", err
		}
		return c.wrap(output), nil
	}

	pkg, err := ParsePackageDir(dir)
	if err != nil {
		return "", err
	}

	return c.CompilePackage(pkg)
}

func (c *Compiler) CompileSource(source string) (string, error) {
	return c.CompileSourceFile("main.go", source)
}

func (c *Compiler) CompileSourceFile(filename, source string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}

	if strings.TrimSpace(filename) == "" {
		filename = "main.go"
	}

	dir, err := os.MkdirTemp("", "go2js-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	filename = filepath.Base(filename)
	if filepath.Ext(filename) != ".go" {
		return "", fmt.Errorf("source filename must have .go extension")
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		return "", err
	}

	return c.CompileFile(path)
}

func (c *Compiler) wrap(source string) string {
	options := javascript.ModuleOptions{
		Format: javascript.ModuleFormat(c.Options.Module),
		Strict: c.Options.Strict,
	}

	output := javascript.WrapModule(source, options)

	if c.Options.Strict && !strings.Contains(output, `"use strict";`) {
		output = `"use strict";\n` + output
	}

	return javascript.FormatJavaScript(output, c.Options.Minify)
}
