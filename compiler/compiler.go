package compiler

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

type Program struct {
	File     *ast.File
	Analysis *Analysis
}

type PackageProgram struct {
	Package  *Package
	Analysis *PackageAnalysis
}

func CompileFile(path string) (string, error) {
	parsed, err := ParseFile(path)
	if err != nil {
		return "", err
	}

	analysis, err := Analyze(parsed)
	if err != nil {
		return "", err
	}

	return javascript.EmitWithContext(parsed.File, analysis.Types, analysis.Semantic)
}

func CompilePackage(pkg *Package) (string, error) {
	analysis, err := AnalyzePackage(pkg)
	if err != nil {
		return "", err
	}

	var output string

	for _, parsed := range pkg.Files {
		if parsed == nil || parsed.File == nil {
			return "", fmt.Errorf("package contains invalid file")
		}

		if output != "" {
			output += "\n"
		}

		code, err := javascript.EmitWithContext(parsed.File, analysis.Types, analysis.Semantic)
		if err != nil {
			return "", err
		}

		output += code
	}

	return output, nil
}

func CompileDirectory(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	pkg := NewPackage(filepath.Base(dir))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".go" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		parsed, err := ParseFile(path)
		if err != nil {
			return "", err
		}

		pkg.Add(parsed)
	}

	if pkg.Empty() {
		return "", fmt.Errorf("no Go files found in %s", dir)
	}

	return CompilePackage(pkg)
}
