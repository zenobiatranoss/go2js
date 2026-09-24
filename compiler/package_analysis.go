package compiler

import (
	semantic "github.com/zenobiatranoss/go2js/compiler/semantic"
	gotypes "github.com/zenobiatranoss/go2js/types"
)

type PackageAnalysis struct {
	Package  *Package
	Types    *gotypes.Result
	Semantic *semantic.Context
}
