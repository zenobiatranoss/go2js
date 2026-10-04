package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/types"
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

	// The comments of a file are read as well as its declarations, since what a
	// declaration asks of the JavaScript it is written as is written in the
	// comment above it.
	parsed, err := ParseFileWithOptions(path, ParseOptions{ParseComments: true})
	if err != nil {
		return "", err
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		return "", err
	}

	exports := ExportedDeclarations([]*ParsedFile{parsed}, analysis.Types)

	output, err := javascript.EmitWithContextOptionsTarget(
		parsed.File,
		analysis.Types,
		analysis.Semantic,
		c.Options.Runtime,
		c.Options.Target,
		Handlers(exports)...,
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

	return c.wrapExportsWithSources(runDeferredInits(output), sources, exports), nil
}

// runDeferredInits makes the assignments that were held back while a package
// level variable was written. Go makes them once every declaration in the
// package is in place, which is before the init functions of the package run and
// before main is called, so the call is placed ahead of both.
func runDeferredInits(code string) string {
	if !strings.Contains(code, "go2jsDeferInit(") {
		return code
	}

	return strings.Replace(code, "return yield* main();", "go2jsRunInitializers();\nreturn yield* main();", 1)
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
	needsRuntime := false
	packagePath := ""

	exports := ExportedDeclarations(pkg.Files, analysis.Types)

	for _, parsed := range pkg.Files {
		if parsed == nil || parsed.File == nil {
			return "", fmt.Errorf("package contains invalid file")
		}

		// The path of the package is told to the emitter so that a type declared in
		// another file of it is known to be one of this package, which a file
		// only knows about its own declarations.
		if packagePath == "" {
			packagePath = packagePathOf(analysis.Types)
		}

		// The runtime is written once for the package rather than once for each
		// file of it, so the files are emitted without it and it is put in front
		// of the whole package if any of them asks for it. What the package hands
		// over is handed over from the file each declaration was written in, which
		// is where the declaration it hands over is.
		code, needs, err := javascript.EmitFile(
			parsed.File,
			analysis.Types,
			analysis.Semantic,
			false,
			c.Options.Target,
			packagePath,
			nil,
			Handlers(fileExportsOf(parsed, analysis.Types))...,
		)
		if err != nil {
			return "", err
		}

		if needs {
			needsRuntime = true
		}

		parts = append(parts, code)
	}

	code := strings.Join(parts, "\n")

	if c.Options.Runtime && needsRuntime {
		code = javascript.ProgramRuntime(code, c.Options.Target) + code
	}

	var sources []javascript.SourceMapSource
	if c.Options.SourceMap {
		sources = packageSourceMapSources(pkg)
	}

	return c.wrapExportsWithSources(runDeferredInits(code), sources, exports), nil
}

func (c *Compiler) CompileDirectory(dir string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil compiler")
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", fmt.Errorf("empty source directory")
	}

	// A project is written out as the module it is written for, with what it was
	// asked to hand over handed over, so it is not wrapped again here.
	if _, _, err := findModuleRoot(dir); err == nil {
		return compileProjectWithOptions(dir, c.Options)
	}

	pkg, err := ParsePackageDirWithOptions(dir, ParseOptions{ParseComments: true})
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

	// A project is written out as the module it is written for, with what it was
	// asked to hand over handed over, so it is not wrapped again here.
	return CompileProjectWithOptions(dir, c.Options)
}

// wrapExportsWithSources writes a program out as the module it is written for,
// with what it exports handed over at the end of it, and with the source it was
// written from beside it when a source map was asked for.
//
// The exports go inside the module rather than after it, since where a module
// names what it exports is inside the module, and a program written as an
// expression hands nothing over because an expression has nowhere to hand
// anything to.
func (c *Compiler) wrapExportsWithSources(source string, sources []javascript.SourceMapSource, exports []Export) string {
	output := wrapExports(source, c.Options, exports)

	if c.Options.SourceMap {
		output = javascript.AppendInlineSourceMap(output, sources)
	}

	return output
}

// wrapExports writes a program out as the module it is written for, with what it
// exports handed over at the end of it.
func wrapExports(source string, options Options, exports []Export) string {
	module := javascript.ModuleOptions{
		Format: javascript.ModuleFormat(options.Module),
		Strict: options.Strict,
	}

	source += ExportHandlers(exports, module.Format)

	output := javascript.WrapModule(source, module)

	if options.Strict && !strings.Contains(output, `"use strict";`) {
		output = `"use strict";` + "\n" + output
	}

	return javascript.FormatJavaScript(output, options.Minify)
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

// packagePathOf is the path the go tool knows the package a file belongs to by,
// which is what tells a declaration of this package apart from one of a package
// beside it. A file compiled on its own carries the path of a whole program
// rather than of a package, and there is nothing to tell apart there, so the
// emitter is told nothing in that case.
func packagePathOf(analysis *types.Result) string {
	if analysis == nil || analysis.Package == nil {
		return ""
	}

	path := analysis.Package.Path()
	if path == "" || path == "command-line-arguments" {
		return ""
	}

	return path
}
