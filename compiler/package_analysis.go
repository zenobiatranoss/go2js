package compiler

import (
	semantic "github.com/zenobiatranoss/GonvertJS/compiler/semantic"
	gotypes "github.com/zenobiatranoss/GonvertJS/types"
)

type PackageAnalysis struct {
	Package  *Package
	Types    *gotypes.Result
	Semantic *semantic.Context
}
