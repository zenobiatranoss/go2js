package compiler

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/types"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/zenobiatranoss/go2js/backend/javascript"
	"github.com/zenobiatranoss/go2js/compiler/semantic"
	gotypes "github.com/zenobiatranoss/go2js/types"
)

type projectPackage struct {
	path     string
	dir      string
	pkg      *Package
	analysis *Analysis
}

type project struct {
	root     string
	module   string
	options  Options
	packages map[string]*projectPackage
	active   map[string]bool
	importer *projectImporter
}

type projectImporter struct {
	project  *project
	fallback types.Importer
}

func (i *projectImporter) Import(path string) (*types.Package, error) {
	if i.project.isLocal(path) {
		pkg, err := i.project.load(path)
		if err != nil {
			return nil, err
		}
		return pkg.analysis.Types.Package, nil
	}

	return i.fallback.Import(path)
}

func compileProjectWithOptions(dir string, options Options) (string, error) {
	root, module, err := findModuleRoot(dir)
	if err != nil {
		return "", err
	}

	targetDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	targetPath, err := modulePathForDir(root, module, targetDir)
	if err != nil {
		return "", err
	}

	p := &project{
		options:  options.Normalize(),
		root:     root,
		module:   module,
		packages: make(map[string]*projectPackage),
		active:   make(map[string]bool),
	}

	p.importer = &projectImporter{
		project:  p,
		fallback: importer.Default(),
	}

	target, err := p.load(targetPath)
	if err != nil {
		return "", err
	}

	if !target.pkg.HasMain() {
		return "", fmt.Errorf("compiler: package %q has no main function", targetPath)
	}

	order, err := p.order(targetPath)
	if err != nil {
		return "", err
	}

	var out strings.Builder

	for _, path := range order {
		code, err := p.emitNamespace(p.packages[path])
		if err != nil {
			return "", err
		}
		if code != "" {
			out.WriteString(code)
			out.WriteString("\n")
		}
	}

	if err := p.emitImports(&out, target.pkg); err != nil {
		return "", err
	}

	mainCode, err := p.emitFiles(target)
	if err != nil {
		return "", err
	}

	mainCode, initCalls := renameInitFunctions(target.pkg, mainCode)
	mainCode = strings.Replace(mainCode, "main();", initCalls+"main();", 1)
	out.WriteString(mainCode)
	return out.String(), nil
}

func (p *project) load(path string) (*projectPackage, error) {
	if loaded := p.packages[path]; loaded != nil {
		return loaded, nil
	}

	if p.active[path] {
		return nil, fmt.Errorf("compiler: import cycle involving %q", path)
	}

	dir, ok := p.localDir(path)
	if !ok {
		return nil, fmt.Errorf("compiler: local import %q cannot be resolved", path)
	}

	p.active[path] = true
	defer delete(p.active, path)

	pkg, err := ParsePackageDir(dir)
	if err != nil {
		return nil, err
	}

	result, err := gotypes.CheckFilesWithConfig(
		pkg.FileSet(),
		pkg.ASTFiles(),
		gotypes.Config{
			Importer:    p.importer,
			PackagePath: path,
		},
	)
	if err != nil {
		return nil, err
	}

	analysis := &Analysis{
		Types:    result,
		Package:  pkg,
		Semantic: semantic.NewResultContext(result, pkg.FileSet()),
	}

	loaded := &projectPackage{
		path:     path,
		dir:      dir,
		pkg:      pkg,
		analysis: analysis,
	}

	p.packages[path] = loaded
	return loaded, nil
}

func (p *project) order(root string) ([]string, error) {
	seen := make(map[string]bool)
	active := make(map[string]bool)
	order := make([]string, 0)

	var visit func(string) error
	visit = func(path string) error {
		if active[path] {
			return fmt.Errorf("compiler: import cycle involving %q", path)
		}
		if seen[path] {
			return nil
		}

		pkg, err := p.load(path)
		if err != nil {
			return err
		}

		active[path] = true

		for _, imported := range pkg.pkg.Imports() {
			if p.isLocal(imported) {
				if err := visit(imported); err != nil {
					return err
				}
			}
		}

		delete(active, path)
		seen[path] = true

		if path != root {
			order = append(order, path)
		}

		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}

	return order, nil
}

func (p *project) emitNamespace(pkg *projectPackage) (string, error) {
	var out strings.Builder
	namespace := p.namespace(pkg.path)

	out.WriteString("const ")
	out.WriteString(namespace)
	out.WriteString(" = (() => {\n")

	for _, imported := range pkg.pkg.Imports() {
		if !p.isLocal(imported) {
			continue
		}

		if p.isDotImport(pkg.pkg, imported) {
			names, err := p.dotImportNames(imported)
			if err != nil {
				return "", err
			}

			for _, name := range names {
				out.WriteString("let ")
				out.WriteString(name)
				out.WriteString(" = ")
				out.WriteString(p.namespace(imported))
				out.WriteString(".")
				out.WriteString(name)
				out.WriteString(";\n")
			}
			continue
		}

		alias, err := p.alias(pkg.pkg, imported)
		if err != nil {
			return "", err
		}

		if alias == "" {
			continue
		}

		out.WriteString("const ")
		out.WriteString(alias)
		out.WriteString(" = ")
		out.WriteString(p.namespace(imported))
		out.WriteString(";\n")
	}

	if len(pkg.pkg.Imports()) > 0 {
		out.WriteString("\n")
	}

	code, err := p.emitFiles(pkg)
	if err != nil {
		return "", err
	}

	code, initCalls := renameInitFunctions(pkg.pkg, code)
	out.WriteString(code)
	out.WriteString(initCalls)
	out.WriteString("\nreturn {")

	names := exportedNames(pkg.pkg)
	sort.Strings(names)

	if len(names) > 0 {
		out.WriteString("\n")
		for i, name := range names {
			out.WriteString("    ")
			out.WriteString(name)
			if i+1 < len(names) {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
	}

	out.WriteString("};\n})();\n")
	return out.String(), nil
}

func (p *project) emitImports(out *strings.Builder, pkg *Package) error {
	for _, imported := range pkg.Imports() {
		if !p.isLocal(imported) {
			continue
		}

		if p.isDotImport(pkg, imported) {
			names, err := p.dotImportNames(imported)
			if err != nil {
				return err
			}

			for _, name := range names {
				out.WriteString("let ")
				out.WriteString(name)
				out.WriteString(" = ")
				out.WriteString(p.namespace(imported))
				out.WriteString(".")
				out.WriteString(name)
				out.WriteString(";\n")
			}
			continue
		}

		alias, err := p.alias(pkg, imported)
		if err != nil {
			return err
		}

		if alias == "" {
			continue
		}

		out.WriteString("const ")
		out.WriteString(alias)
		out.WriteString(" = ")
		out.WriteString(p.namespace(imported))
		out.WriteString(";\n")
	}

	if len(pkg.Imports()) > 0 {
		out.WriteString("\n")
	}

	return nil
}

func (p *project) emitFiles(pkg *projectPackage) (string, error) {
	var out strings.Builder

	for _, file := range pkg.pkg.Files {
		if file == nil || file.File == nil {
			return "", fmt.Errorf("compiler: package %q contains invalid file", pkg.path)
		}

		code, err := javascript.EmitWithContextOptionsTarget(
			file.File,
			pkg.analysis.Types,
			pkg.analysis.Semantic,
			p.options.Runtime,
			p.options.Target,
		)
		if err != nil {
			return "", err
		}

		if out.Len() > 0 {
			out.WriteString("\n")
		}

		out.WriteString(code)
	}

	return out.String(), nil
}

func (p *project) alias(pkg *Package, path string) (string, error) {
	for _, file := range pkg.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, spec := range file.File.Imports {
			if spec.Path == nil {
				continue
			}

			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return "", err
			}

			if value != path {
				continue
			}

			if spec.Name != nil {
				switch spec.Name.Name {
				case "_":
					return "", nil
				case ".":
					return "", nil
				default:
					return spec.Name.Name, nil
				}
			}

			loaded, err := p.load(path)
			if err != nil {
				return "", err
			}

			return loaded.pkg.Name, nil
		}
	}

	return "", fmt.Errorf("compiler: import %q not found", path)
}

func (p *project) isDotImport(pkg *Package, path string) bool {
	if pkg == nil {
		return false
	}

	for _, file := range pkg.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, spec := range file.File.Imports {
			if spec.Path == nil {
				continue
			}

			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil || value != path || spec.Name == nil {
				continue
			}

			return spec.Name.Name == "."
		}
	}

	return false
}

func (p *project) dotImportNames(path string) ([]string, error) {
	loaded, err := p.load(path)
	if err != nil {
		return nil, err
	}

	names := exportedNames(loaded.pkg)
	sort.Strings(names)
	return names, nil
}

func (p *project) isLocal(path string) bool {
	return path == p.module || strings.HasPrefix(path, p.module+"/")
}

func (p *project) localDir(path string) (string, bool) {
	if !p.isLocal(path) {
		return "", false
	}

	if path == p.module {
		return p.root, true
	}

	rel := strings.TrimPrefix(path, p.module+"/")
	dir := filepath.Clean(filepath.Join(p.root, filepath.FromSlash(rel)))

	relative, err := filepath.Rel(p.root, dir)
	if err != nil {
		return "", false
	}

	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", false
	}

	return dir, true
}

func (p *project) namespace(path string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(path))
	return fmt.Sprintf("go2js_pkg_%08x", h.Sum32())
}

func findModuleRoot(start string) (string, string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", "", err
	}

	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			module, err := parseModulePath(data)
			if err != nil {
				return "", "", err
			}
			return dir, module, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", "", fmt.Errorf("compiler: go.mod not found from %s", start)
}

func parseModulePath(data []byte) (string, error) {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
	}

	return "", fmt.Errorf("compiler: module directive not found")
}

func modulePathForDir(root, module, dir string) (string, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return "", err
	}

	if rel == "." {
		return module, nil
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("compiler: directory is outside module")
	}

	return module + "/" + filepath.ToSlash(rel), nil
}

func exportedNames(pkg *Package) []string {
	seen := make(map[string]bool)
	names := make([]string, 0)

	for _, file := range pkg.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, decl := range file.File.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil && d.Name != nil && ast.IsExported(d.Name.Name) && !seen[d.Name.Name] {
					seen[d.Name.Name] = true
					names = append(names, d.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name != nil && ast.IsExported(s.Name.Name) && !seen[s.Name.Name] {
							seen[s.Name.Name] = true
							names = append(names, s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name != nil && ast.IsExported(name.Name) && !seen[name.Name] {
								seen[name.Name] = true
								names = append(names, name.Name)
							}
						}
					}
				}
			}
		}
	}

	return names
}
func renameInitFunctions(pkg *Package, code string) (string, string) {
	calls := strings.Builder{}
	index := 0
	for _, file := range pkg.Files {
		if file == nil || file.File == nil {
			continue
		}
		for _, decl := range file.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name == nil || fn.Name.Name != "init" {
				continue
			}
			name := fmt.Sprintf("go2js_init_%d", index)
			code = strings.Replace(code, "function init(", "function "+name+"(", 1)
			calls.WriteString(name + "();\n")
			index++
		}
	}
	return code, calls.String()
}

func CompileProject(dir string) (string, error) {
	return CompileProjectWithOptions(dir, DefaultOptions())
}

func CompileProjectWithOptions(dir string, options Options) (string, error) {
	options = options.Normalize()
	if err := options.Validate(); err != nil {
		return "", err
	}

	output, err := compileProjectWithOptions(dir, options)
	if err != nil {
		return "", err
	}

	return javascript.FormatJavaScript(output, options.Minify), nil
}
