package compiler

import (
	"github.com/zenobiatranoss/go2js/types"
)

type Analysis struct {
	Types *types.Result
}

func Analyze(parsed *ParsedFile) (*Analysis, error) {
	result, err := types.Check(parsed.Fset, parsed.File)
	if err != nil {
		return nil, err
	}

	return &Analysis{
		Types: result,
	}, nil
}
