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

	output, err := javascript.EmitWithContextOptionsTarget(
		parsed.File,
		analysis.Types,
		analysis.Semantic,
		c.Options.Runtime,
		c.Options.Target,
	)
	if err != nil {
		return "", err
	}

	var sources []javascript.SourceMapSource
	if c.Options.SourceMap {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", readErr
		}

		sources = []javascript.SourceMapSource{{
			Name:    filepath.Base(path),
			Content: string(data),
		}}
	}

	return c.wrapWithSources(output, sources), nil
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

		code, err := javascript.EmitWithContextOptionsTarget(
			parsed.File,
			analysis.Types,
			analysis.Semantic,
			c.Options.Runtime,
			c.Options.Target,
		)
		if err != nil {
			return "", err
		}

		parts = append(parts, code)
	}

	var sources []javascript.SourceMapSource
	if c.Options.SourceMap {
		sources = packageSourceMapSources(pkg)
	}

	return c.wrapWithSources(strings.Join(parts, "\n"), sources), nil
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

func (c *Compiler) CompileProject(dir string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("empty project directory")
	}

	if _, _, err := findModuleRoot(dir); err != nil {
		return "", err
	}

	output, err := CompileProjectWithOptions(dir, c.Options)
	if err != nil {
		return "", err
	}

	return c.wrap(output), nil
}

func (c *Compiler) wrap(source string) string {
	return c.wrapWithSources(source, nil)
}

func (c *Compiler) wrapWithSources(source string, sources []javascript.SourceMapSource) string {
	options := javascript.ModuleOptions{
		Format: javascript.ModuleFormat(c.Options.Module),
		Strict: c.Options.Strict,
	}

	output := javascript.WrapModule(source, options)

	if c.Options.Strict && !strings.Contains(output, `"use strict";`) {
		output = `"use strict";\n` + output
	}

	output = javascript.FormatJavaScript(output, c.Options.Minify)
	if c.Options.SourceMap {
		output = javascript.AppendInlineSourceMap(output, sources)
	}
	return output
}

func packageSourceMapSources(pkg *Package) []javascript.SourceMapSource {
	if pkg == nil {
		return nil
	}

	fset := pkg.FileSet()
	sources := make([]javascript.SourceMapSource, 0, len(pkg.Files))
	seen := make(map[string]bool)

	for _, parsed := range pkg.Files {
		if parsed == nil || parsed.File == nil {
			continue
		}

		filename := fset.Position(parsed.File.Pos()).Filename
		if filename == "" || seen[filename] {
			continue
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		seen[filename] = true
		sources = append(sources, javascript.SourceMapSource{
			Name:    filepath.Base(filename),
			Content: string(data),
		})
	}

	return sources
}
