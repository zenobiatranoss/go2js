package compiler

import (
	"go/ast"
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
	return NewDefault().CompileFile(path)
}

func CompilePackage(pkg *Package) (string, error) {
	return NewDefault().CompilePackage(pkg)
}

func CompileDirectory(dir string) (string, error) {
	return NewDefault().CompileDirectory(dir)
}
