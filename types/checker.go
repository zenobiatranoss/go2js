package types

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	gotypes "go/types"
)

type Config struct {
	Importer         gotypes.Importer
	GoVersion        string
	Sizes            gotypes.Sizes
	PackagePath      string
	IgnoreFuncBodies bool
	Error            func(error)
}

func (c Config) std() *gotypes.Config {
	config := &gotypes.Config{
		Importer:         c.Importer,
		GoVersion:        c.GoVersion,
		Sizes:            c.Sizes,
		IgnoreFuncBodies: c.IgnoreFuncBodies,
		Error:            c.Error,
	}

	if config.Importer == nil {
		config.Importer = importer.Default()
	}

	return config
}

func Check(fset *token.FileSet, file *ast.File) (*Result, error) {
	return CheckFilesWithConfig(fset, []*ast.File{file}, Config{})
}

func CheckFiles(fset *token.FileSet, files []*ast.File) (*Result, error) {
	return CheckFilesWithConfig(fset, files, Config{})
}

func CheckWithConfig(fset *token.FileSet, file *ast.File, config Config) (*Result, error) {
	return CheckFilesWithConfig(fset, []*ast.File{file}, config)
}

func CheckFilesWithConfig(fset *token.FileSet, files []*ast.File, config Config) (*Result, error) {
	if fset == nil {
		return nil, fmt.Errorf("nil file set")
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files to check")
	}

	info := &gotypes.Info{
		Types:        make(map[ast.Expr]gotypes.TypeAndValue),
		Defs:         make(map[*ast.Ident]gotypes.Object),
		Uses:         make(map[*ast.Ident]gotypes.Object),
		Implicits:    make(map[ast.Node]gotypes.Object),
		Instances:    make(map[*ast.Ident]gotypes.Instance),
		Selections:   make(map[*ast.SelectorExpr]*gotypes.Selection),
		Scopes:       make(map[ast.Node]*gotypes.Scope),
		FileVersions: make(map[*ast.File]string),
	}

	pkgName := ""
	for _, file := range files {
		if file == nil || file.Name == nil {
			return nil, fmt.Errorf("invalid file")
		}

		if pkgName == "" {
			pkgName = file.Name.Name
			continue
		}

		if file.Name.Name != pkgName {
			return nil, fmt.Errorf(
				"files belong to different packages: %s and %s",
				pkgName,
				file.Name.Name,
			)
		}
	}

	packagePath := config.PackagePath
	if packagePath == "" {
		packagePath = pkgName
	}

	pkg, err := config.std().Check(packagePath, fset, files, info)
	if err != nil {
		return nil, err
	}

	return &Result{
		Info:         info,
		Types:        info.Types,
		Defs:         info.Defs,
		Uses:         info.Uses,
		Implicits:    info.Implicits,
		Instances:    info.Instances,
		Selections:   info.Selections,
		Scopes:       info.Scopes,
		InitOrder:    info.InitOrder,
		FileVersions: info.FileVersions,
		Package:      pkg,
	}, nil
}
