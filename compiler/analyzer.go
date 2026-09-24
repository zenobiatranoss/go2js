package compiler

import (
	"fmt"
	semantic "github.com/zenobiatranoss/go2js/compiler/semantic"

	"github.com/zenobiatranoss/go2js/types"
)

type Analysis struct {
	Types    *types.Result
	Package  *Package
	Semantic *semantic.Context
}

func Analyze(parsed *ParsedFile) (*Analysis, error) {
	if parsed == nil || !parsed.Valid() {
		return nil, fmt.Errorf("compiler: invalid parsed file")
	}

	result, err := types.Check(parsed.Fset, parsed.File)
	if err != nil {
		return nil, err
	}

	return &Analysis{
		Types:    result,
		Package:  nil,
		Semantic: semantic.NewResultContext(result, parsed.Fset),
	}, nil
}

func AnalyzePackage(pkg *Package) (*Analysis, error) {
	if err := pkg.Validate(); err != nil {
		return nil, err
	}

	result, err := types.CheckFiles(pkg.FileSet(), pkg.ASTFiles())
	if err != nil {
		return nil, err
	}

	return &Analysis{
		Types:    result,
		Package:  pkg,
		Semantic: semantic.NewResultContext(result, pkg.FileSet()),
	}, nil
}
