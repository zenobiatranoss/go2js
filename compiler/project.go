package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"hash/fnv"
	"os"
	"os/exec"
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
	root           string
	module         string
	options        Options
	packages       map[string]*projectPackage
	active         map[string]bool
	dependencyDirs map[string]string
	importer       *projectImporter
	needsRuntime   bool
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

	if _, ok := i.project.dependencyDir(path); ok {
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
		options:        options.Normalize(),
		root:           root,
		module:         module,
		packages:       make(map[string]*projectPackage),
		active:         make(map[string]bool),
		dependencyDirs: make(map[string]string),
	}

	p.importer = &projectImporter{
		project:  p,
		fallback: newModuleImporter(root),
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

	mainCode, err := p.emitFiles(target, true)
	if err != nil {
		return "", err
	}

	mainCode, initCalls := renameInitFunctions(target.pkg, mainCode)

	// A package level variable is assigned once every declaration in the package
	// is in place, which is before the init functions of the package run and
	// before main is called. A variable declared in a file other than the one
	// holding main is why this is made here rather than beside the call.
	if strings.Contains(mainCode, "go2jsDeferInit(") {
		initCalls = "go2jsRunInitializers();\n" + initCalls
	}

	mainCode = strings.Replace(mainCode, "return yield* main();", initCalls+"return yield* main();", 1)
	out.WriteString(mainCode)

	program := out.String()

	if p.options.Runtime && p.needsRuntime {
		program = javascript.ProgramRuntime(program, p.options.Target) + program
	}

	// What the program hands over is handed over from the package that was asked
	// for, which is the one the program is written around.
	exports := ExportedDeclarations(target.pkg.Files, target.analysis.Types)

	return wrapExports(program, p.options, exports), nil
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
		dir, ok = p.dependencyDir(path)
		if !ok {
			return nil, fmt.Errorf("compiler: local import %q cannot be resolved", path)
		}
	}

	p.active[path] = true
	defer delete(p.active, path)

	// The comments of a file are read as well as its declarations, since what a
	// declaration asks of the JavaScript it is written as is written in the
	// comment above it.
	pkg, err := ParsePackageDirWithOptions(dir, ParseOptions{ParseComments: true})
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
			if !p.isTranspilable(imported) {
				continue
			}

			if err := visit(imported); err != nil {
				return err
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

	emittedNames := map[string]bool{}

	for _, imported := range pkg.pkg.Imports() {
		if !p.isTranspilable(imported) {
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

		for _, alias := range p.localNames(pkg.pkg)[imported] {
			// An alias is a declaration, so it takes the same JavaScript spelling
			// every use of it is written with.
			alias = javascript.Identifier(alias)

			if emittedNames[alias] {
				continue
			}

			emittedNames[alias] = true

			out.WriteString("const ")
			out.WriteString(alias)
			out.WriteString(" = ")
			out.WriteString(p.namespace(imported))
			out.WriteString(";\n")
		}
	}

	if len(pkg.pkg.Imports()) > 0 {
		out.WriteString("\n")
	}

	code, err := p.emitFiles(pkg, false)
	if err != nil {
		return "", err
	}

	code, initCalls := renameInitFunctions(pkg.pkg, code)
	out.WriteString(code)

	// A package level variable is assigned once every declaration in the package
	// is in place, which is before the init functions of that package run, so
	// the assignments that were held back are made here. The init functions are
	// Go functions, which are generators, and a namespace is not one, so they
	// are run through rather than stepped from here.
	if strings.Contains(code, "go2jsDeferInit(") {
		out.WriteString("go2jsRunInitializers();\n")
	}

	if initCalls != "" {
		out.WriteString("go2jsRunSync(function* () {\n")
		out.WriteString(initCalls)
		out.WriteString("});\n")
	}

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
	emitted := map[string]bool{}

	for _, imported := range pkg.Imports() {
		if !p.isTranspilable(imported) {
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

		for _, alias := range p.localNames(pkg)[imported] {
			alias = javascript.Identifier(alias)

			if emitted[alias] {
				continue
			}

			emitted[alias] = true

			out.WriteString("const ")
			out.WriteString(alias)
			out.WriteString(" = ")
			out.WriteString(p.namespace(imported))
			out.WriteString(";\n")
		}
	}

	if len(pkg.Imports()) > 0 {
		out.WriteString("\n")
	}

	return nil
}

func moveMainFileLast(ordered []*ParsedFile) []*ParsedFile {
	for index, file := range ordered {
		if file == nil || !hasMainFunc(file.File) {
			continue
		}

		if index == len(ordered)-1 {
			return ordered
		}

		result := make([]*ParsedFile, 0, len(ordered))
		result = append(result, ordered[:index]...)
		result = append(result, ordered[index+1:]...)
		result = append(result, file)

		return result
	}

	return ordered
}

// emitFiles writes the files of a package out. What the program hands over is
// only handed over from the package the program is written around, so the
// declarations of any other package are written as they stand: a declaration
// asked to be handed over is only asked to be handed over by the program it is
// part of, and inside another package it is nothing but a declaration.
func (p *project) emitFiles(pkg *projectPackage, exposed bool) (string, error) {
	var out strings.Builder

	qualifiers := map[string]string{}

	locals := p.localNames(pkg.pkg)

	for path, list := range locals {
		if len(list) == 0 {
			continue
		}

		qualifiers[path] = list[0]
	}

	// A dot import brings its names in unqualified, so there is no qualifier to
	// write, but the package is still one this compilation writes out, which is
	// what an empty one records.
	for _, imported := range pkg.pkg.Imports() {
		if !p.isTranspilable(imported) {
			continue
		}

		if _, ok := qualifiers[imported]; ok {
			continue
		}

		if p.isDotImport(pkg.pkg, imported) {
			qualifiers[imported] = ""
		}
	}

	ordered := moveMainFileLast(orderPackageFiles(pkg.pkg))

	passes := []func(*ast.File, *gotypes.Result, *semantic.Context, bool, string, string, map[string]string, ...javascript.Export) (string, bool, error){
		javascript.EmitFileTypes,
		javascript.EmitFileBody,
	}

	for _, emit := range passes {
		for _, file := range ordered {
			var handlers []javascript.Export

			if exposed {
				handlers = Handlers(fileExports(file.File, pkg.analysis.Types))
			}

			code, needsRuntime, err := emit(
				file.File,
				pkg.analysis.Types,
				pkg.analysis.Semantic,
				false,
				p.options.Target,
				pkg.path,
				qualifiers,
				handlers...,
			)
			if err != nil {
				return "", err
			}

			if needsRuntime {
				p.needsRuntime = true
			}

			if code == "" {
				continue
			}

			if out.Len() > 0 {
				out.WriteString("\n")
			}

			out.WriteString(code)
		}
	}

	return out.String(), nil
}

func hasMainFunc(file *ast.File) bool {
	if file == nil {
		return false
	}

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name != nil && fn.Name.Name == "main" {
			return true
		}
	}

	return false
}

func (p *project) localNames(pkg *Package) map[string][]string {
	names := map[string]map[string]bool{}

	for _, file := range pkg.Files {
		if file == nil || file.File == nil {
			continue
		}

		for _, spec := range file.File.Imports {
			if spec.Path == nil {
				continue
			}

			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil || path == "" {
				continue
			}

			if !p.isTranspilable(path) {
				continue
			}

			local := ""

			switch {
			case spec.Name == nil:
				loaded, err := p.load(path)
				if err != nil || loaded.pkg.Name == "" {
					continue
				}

				local = loaded.pkg.Name
			case spec.Name.Name == "_" || spec.Name.Name == ".":
				continue
			default:
				local = spec.Name.Name
			}

			if names[path] == nil {
				names[path] = map[string]bool{}
			}

			names[path][local] = true
		}
	}

	result := map[string][]string{}

	for path, set := range names {
		list := make([]string, 0, len(set))

		for name := range set {
			list = append(list, name)
		}

		sort.Strings(list)
		result[path] = list
	}

	return result
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

func (p *project) isTranspilable(path string) bool {
	if p.isLocal(path) {
		return true
	}

	_, ok := p.dependencyDir(path)

	return ok
}

func (p *project) dependencyDir(path string) (string, bool) {
	if path == "" || !strings.Contains(path, ".") {
		return "", false
	}

	if dir, ok := p.dependencyDirs[path]; ok {
		return dir, dir != ""
	}

	dir := p.lookupDependencyDir(path)
	p.dependencyDirs[path] = dir

	return dir, dir != ""
}

func (p *project) lookupDependencyDir(path string) string {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", path)
	cmd.Dir = p.root

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	dir := strings.TrimSpace(string(output))
	if dir == "" {
		return ""
	}

	if !strings.HasPrefix(dir, p.root) {
		return dir
	}

	return ""
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
						if d.Tok == token.TYPE && emitsRuntimeValue(s) && s.Name != nil &&
							ast.IsExported(s.Name.Name) && !seen[s.Name.Name] {
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
			code = strings.Replace(code, "function* init(", "function* "+name+"(", 1)
			calls.WriteString("\tyield* " + name + "();\n")
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

func orderPackageFiles(pkg *Package) []*ParsedFile {
	files := make([]*ParsedFile, 0, len(pkg.Files))
	owner := make(map[string]int)
	references := make(map[string]map[string]bool)
	bodies := make(map[string]map[string]bool)

	for index, file := range pkg.Files {
		if file == nil || file.File == nil {
			files = append(files, file)
			continue
		}

		files = append(files, file)
		collectPackageDependencies(file.File, index, owner, references, bodies)
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make([]int, len(files))
	ordered := make([]*ParsedFile, 0, len(files))

	var resolve func(name string, seen map[string]bool) map[string]bool

	resolve = func(name string, seen map[string]bool) map[string]bool {
		if seen[name] {
			return nil
		}

		seen[name] = true

		direct := references[name]
		resolved := make(map[string]bool, len(direct))

		for reference := range direct {
			if _, isVariable := owner[reference]; isVariable && reference != name {
				resolved[reference] = true
				continue
			}

			if _, isFunction := bodies[reference]; isFunction {
				references[reference] = bodies[reference]
			}

			for nested := range resolve(reference, seen) {
				resolved[nested] = true
			}
		}

		return resolved
	}

	uses := make([]map[string]bool, len(files))

	for index, file := range files {
		if file == nil || file.File == nil {
			continue
		}

		uses[index] = make(map[string]bool)

		for name := range packageLevelVarNames(file.File) {
			for dependency := range resolve(name, map[string]bool{}) {
				uses[index][dependency] = true
			}
		}
	}

	typeOwner := collectTypeDependencies(files, uses)

	var visit func(index int)

	visit = func(index int) {
		if index < 0 || index >= len(files) || state[index] != unvisited {
			return
		}

		state[index] = visiting

		for name := range uses[index] {
			declared, ok := typeOwner[name]
			if !ok {
				declared, ok = owner[name]
			}

			if !ok || declared == index {
				continue
			}

			visit(declared)
		}

		state[index] = visited
		ordered = append(ordered, files[index])
	}

	for index := range files {
		visit(index)
	}

	return ordered
}

func collectTypeDependencies(files []*ParsedFile, uses []map[string]bool) map[string]int {
	declared := make(map[string]int)

	for index, file := range files {
		if file == nil || file.File == nil {
			continue
		}

		for name := range packageLevelTypeNames(file.File) {
			declared[name] = index
		}
	}

	for index, file := range files {
		if file == nil || file.File == nil {
			continue
		}

		if uses[index] == nil {
			uses[index] = make(map[string]bool)
		}

		for _, decl := range file.File.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Name == nil {
				continue
			}

			receiver := methodReceiverTypeName(function)

			if receiver == "" {
				continue
			}

			if target, ok := declared[receiver]; ok && target != index {
				uses[index][receiver] = true
				_ = target
			}

			if function.Body == nil {
				continue
			}

			for name := range referencedIdents(function.Body) {
				if target, ok := declared[name]; ok && target != index {
					uses[index][name] = true
				}
			}

			for _, parameter := range function.Type.Params.List {
				for name := range referencedIdents(parameter.Type) {
					if target, ok := declared[name]; ok && target != index {
						uses[index][name] = true
					}
				}
			}
		}
	}

	return declared
}

func methodReceiverTypeName(function *ast.FuncDecl) string {
	if function == nil || function.Recv == nil || len(function.Recv.List) != 1 {
		return ""
	}

	return receiverTypeName(function.Recv.List[0].Type)
}

func receiverTypeName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverTypeName(value.X)
	case *ast.IndexExpr:
		return receiverTypeName(value.X)
	case *ast.IndexListExpr:
		return receiverTypeName(value.X)
	}

	return ""
}

func packageLevelTypeNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			if typed, ok := spec.(*ast.TypeSpec); ok && typed.Name != nil {
				names[typed.Name.Name] = true
			}
		}
	}

	return names
}

func collectPackageDependencies(
	file *ast.File,
	index int,
	owner map[string]int,
	references map[string]map[string]bool,
	bodies map[string]map[string]bool,
) {
	for name := range packageLevelVarNames(file) {
		owner[name] = index
	}

	for _, decl := range file.Decls {
		switch declaration := decl.(type) {
		case *ast.FuncDecl:
			if declaration.Recv != nil || declaration.Name == nil {
				continue
			}

			if _, seen := bodies[declaration.Name.Name]; !seen {
				bodies[declaration.Name.Name] = make(map[string]bool)
			}

			declaredNames := declaration.Name.Name

			if declaration.Body != nil {
				for name := range referencedIdents(declaration.Body) {
					bodies[declaredNames][name] = true
				}
			}

		case *ast.GenDecl:
			if declaration.Tok != token.VAR {
				continue
			}

			for _, spec := range declaration.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				used := make(map[string]bool)

				for _, expr := range value.Values {
					for name := range referencedIdents(expr) {
						used[name] = true
					}
				}

				for _, name := range value.Names {
					if name != nil {
						references[name.Name] = used
					}
				}
			}
		}
	}
}

func referencedIdents(node ast.Node) map[string]bool {
	found := make(map[string]bool)

	ast.Inspect(node, func(current ast.Node) bool {
		if ident, ok := current.(*ast.Ident); ok {
			found[ident.Name] = true
		}

		return true
	})

	return found
}

func packageLevelVarNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range value.Names {
				if name != nil {
					names[name.Name] = true
				}
			}
		}
	}

	return names
}

func initializerReferences(file *ast.File) map[string]bool {
	references := make(map[string]bool)

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, expr := range value.Values {
				ast.Inspect(expr, func(node ast.Node) bool {
					if ident, ok := node.(*ast.Ident); ok {
						references[ident.Name] = true
					}

					return true
				})
			}
		}
	}

	return references
}
func emitsRuntimeValue(spec *ast.TypeSpec) bool {
	switch declared := spec.Type.(type) {
	case *ast.StructType:
		return true
	case *ast.InterfaceType:
		return declared.Methods != nil && len(declared.Methods.List) > 0
	}

	return false
}
