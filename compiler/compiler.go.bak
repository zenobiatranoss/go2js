package compiler

import (
	"fmt"
	"go/ast"

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
	if _, _, err := findModuleRoot(dir); err == nil {
		return CompileProject(dir)
	}

	pkg, err := ParsePackageDir(dir)
	if err != nil {
		return "", err
	}

	return CompilePackage(pkg)
}
