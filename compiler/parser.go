package compiler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ParsedFile struct {
	File   *ast.File
	Fset   *token.FileSet
	Path   string
	Source []byte
}

type ParseOptions struct {
	Trace                bool
	ParseComments        bool
	SkipObjectResolution bool
}

func parserMode(options ParseOptions) parser.Mode {
	mode := parser.AllErrors

	if options.ParseComments {
		mode |= parser.ParseComments
	}

	if options.SkipObjectResolution {
		mode |= parser.SkipObjectResolution
	}

	return mode
}

func ParseFile(filename string) (*ParsedFile, error) {
	return ParseFileWithOptions(filename, ParseOptions{})
}

func ParseFileWithOptions(filename string, options ParseOptions) (*ParsedFile, error) {
	fset := token.NewFileSet()
	return ParseFileWithFileSet(fset, filename, options)
}

func ParseFileWithFileSet(fset *token.FileSet, filename string, options ParseOptions) (*ParsedFile, error) {
	if fset == nil {
		return nil, fmt.Errorf("compiler: nil file set")
	}

	file, err := parser.ParseFile(fset, filename, nil, parserMode(options))
	if err != nil {
		return nil, err
	}

	source, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		File:   file,
		Fset:   fset,
		Path:   filename,
		Source: source,
	}, nil
}

func ParseSource(filename string, source []byte) (*ParsedFile, error) {
	return ParseSourceWithOptions(filename, source, ParseOptions{})
}

func ParseSourceWithOptions(filename string, source []byte, options ParseOptions) (*ParsedFile, error) {
	fset := token.NewFileSet()
	return ParseSourceWithFileSet(fset, filename, source, options)
}

func ParseSourceWithFileSet(fset *token.FileSet, filename string, source []byte, options ParseOptions) (*ParsedFile, error) {
	if fset == nil {
		return nil, fmt.Errorf("compiler: nil file set")
	}

	file, err := parser.ParseFile(fset, filename, source, parserMode(options))
	if err != nil {
		return nil, err
	}

	return &ParsedFile{
		File:   file,
		Fset:   fset,
		Path:   filename,
		Source: append([]byte(nil), source...),
	}, nil
}

func ParsePackage(filenames []string) (*Package, error) {
	return ParsePackageWithOptions(filenames, ParseOptions{})
}

func ParsePackageWithOptions(filenames []string, options ParseOptions) (*Package, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("compiler: no package files")
	}

	paths := append([]string(nil), filenames...)
	sort.Strings(paths)

	fset := token.NewFileSet()
	pkg := NewPackage("")

	for _, filename := range paths {
		parsed, err := ParseFileWithFileSet(fset, filename, options)
		if err != nil {
			return nil, err
		}

		if pkg.Name == "" {
			pkg.Name = parsed.PackageName()
		}

		if parsed.PackageName() != pkg.Name {
			return nil, fmt.Errorf(
				"compiler: package name mismatch: %s declares %q, expected %q",
				filename,
				parsed.PackageName(),
				pkg.Name,
			)
		}

		pkg.Add(parsed)
	}

	pkg.SetFileSet(fset)
	pkg.Sort()

	return pkg, nil
}

func ParsePackageDir(dirname string) (*Package, error) {
	return ParsePackageDirWithOptions(dirname, ParseOptions{})
}

func ParsePackageDirWithOptions(dirname string, options ParseOptions) (*Package, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".go") {
			continue
		}

		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		files = append(files, filepath.Join(dirname, name))
	}

	sort.Strings(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("compiler: no Go source files in %s", dirname)
	}

	return ParsePackageWithOptions(files, options)
}

func (p *ParsedFile) Position(pos token.Pos) token.Position {
	if p == nil || p.Fset == nil {
		return token.Position{}
	}

	return p.Fset.Position(pos)
}

func (p *ParsedFile) Offset(pos token.Pos) int {
	if p == nil || p.Fset == nil {
		return -1
	}

	position := p.Fset.PositionFor(pos, false)
	return position.Offset
}

func (p *ParsedFile) Valid() bool {
	return p != nil && p.File != nil && p.Fset != nil
}

func (p *ParsedFile) PackageName() string {
	if !p.Valid() || p.File.Name == nil {
		return ""
	}

	return p.File.Name.Name
}

func (p *ParsedFile) Imports() []*ast.ImportSpec {
	if !p.Valid() {
		return nil
	}

	return p.File.Imports
}

func (p *ParsedFile) Declarations() []ast.Decl {
	if !p.Valid() {
		return nil
	}

	return p.File.Decls
}
