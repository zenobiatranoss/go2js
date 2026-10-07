package javascript

import (
	"fmt"
	"go/ast"
	gotypesstd "go/types"
	"sort"
	"strings"
)

var supportedStdlibPackages = func() map[string]map[string]bool {
	_ = stdlibFuncMapsInitialized

	supported := make(map[string]map[string]bool, len(stdlibFuncMaps))

	for name, functions := range stdlibFuncMaps {
		supported[name] = funcSet(functions)
	}

	supported["fmt"] = funcSet(fmtFuncs)
	supported["errors"] = funcSet(errorsFuncs)
	supported["encoding"] = funcSet(encodingFuncs)
	supported["context"] = funcSet(contextFunctions)
	supported["context"] = funcSet(map[string]string{
		"Background":       "go2jsContextBackground",
		"TODO":             "go2jsContextBackground",
		"WithCancel":       "go2jsContextWithCancel",
		"WithValue":        "go2jsContextWithValue",
		"WithTimeout":      "go2jsContextWithTimeout",
		"WithDeadline":     "go2jsContextWithDeadline",
		"AfterFunc":        "go2jsContextAfterFunc",
		"Cause":            "go2jsContextCause",
		"Canceled":         "go2jsContextCanceled",
		"DeadlineExceeded": "go2jsContextDeadlineExceeded",
	})

	for path, aliases := range stdlibPkgAliases {
		for _, alias := range aliases {
			supported[alias] = supported[path]
		}
	}

	return supported
}()

func funcSet(entries map[string]string) map[string]bool {
	set := make(map[string]bool, len(entries))

	for name := range entries {
		set[name] = true
	}

	return set
}

func (e *emitter) packageSelectorName(selector *ast.SelectorExpr) (string, string, bool) {
	if !e.isPackageSelector(selector) {
		return "", "", false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}

	if e.analysis == nil || e.analysis.Uses == nil {
		return "", "", false
	}

	if _, ok := e.analysis.Uses[pkg].(*gotypesstd.PkgName); !ok {
		return "", "", false
	}

	return pkg.Name, selector.Sel.Name, true
}

func (e *emitter) positionOf(node ast.Node) string {
	if e.semantic == nil || node == nil {
		return ""
	}

	position := e.semantic.Position(node.Pos())
	if position.Filename == "" {
		return ""
	}

	return position.String()
}

func (e *emitter) unsupportedStdlibError(position string, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)

	if position == "" {
		return fmt.Errorf("%s", message)
	}

	return fmt.Errorf("%s: %s", position, message)
}

// packageImportPath reports the import path a package selector refers to.
func (e *emitter) packageImportPath(selector *ast.SelectorExpr) (string, bool) {
	if !e.isPackageSelector(selector) {
		return "", false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	if e.analysis == nil || e.analysis.Uses == nil {
		return "", false
	}

	name, ok := e.analysis.Uses[pkg].(*gotypesstd.PkgName)
	if !ok {
		return "", false
	}

	return name.Imported().Path(), true
}

// isStandardLibraryPath reports whether an import path belongs to the standard
// library. Standard library paths have no dot in their first element, while
// module and local package paths are domain qualified.
func isStandardLibraryPath(path string) bool {
	if path == "" {
		return false
	}

	first := path
	if index := strings.Index(path, "/"); index >= 0 {
		first = path[:index]
	}

	return !strings.Contains(first, ".")
}

func (e *emitter) checkSupportedPackageSelector(selector *ast.SelectorExpr) error {
	pkg, name, ok := e.packageSelectorName(selector)
	if !ok {
		return nil
	}

	// Local and third-party packages are not part of the standard library
	// surface this check governs.
	if path, resolved := e.packageImportPath(selector); resolved && !isStandardLibraryPath(path) {
		return nil
	}

	position := e.positionOf(selector)

	key := pkg

	// Two standard library packages can share a local identifier, so prefer
	// the resolved import path over the identifier alone.
	if path, resolved := e.packageImportPath(selector); resolved {
		if _, known := supportedStdlibPackages[path]; known {
			key = path
		}
	}

	functions, supported := supportedStdlibPackages[key]
	if !supported {
		return e.unsupportedStdlibError(position,
			"package %q is not available in the JavaScript standard library; supported packages: %s",
			pkg, strings.Join(knownStdlibPackageNames(), ", "),
		)
	}

	if packageConstants[key+"."+name] != "" || packageTypes[key+"."+name] != "" {
		return nil
	}

	if functions[name] {
		return nil
	}

	return e.unsupportedStdlibError(position,
		"function %s.%s is not available in the JavaScript standard library", pkg, name)
}

func knownStdlibPackageNames() []string {
	names := make([]string, 0, len(supportedStdlibPackages))

	for name := range supportedStdlibPackages {
		if isStdlibPkgAlias(name) {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// SupportedStdlibPackageFuncs reports the symbols the runtime answers for,
// keyed by the name the registry knows them under (mostly import paths). This
// is the inventory the documentation and the gap-report tool ask for, so it
// lives here next to the maps it describes.
func SupportedStdlibPackageFuncs() map[string][]string {
	inventory := make(map[string][]string, len(supportedStdlibPackages))

	for pkg, symbols := range supportedStdlibPackages {
		list := make([]string, 0, len(symbols))
		for name := range symbols {
			list = append(list, name)
		}
		sort.Strings(list)
		inventory[pkg] = list
	}

	return inventory
}

// SupportedStdlibSymbols reports every name the runtime answers for, which is
// the functions of the registry together with the constants and the types it
// writes real code for, grouped by the name the registry knows each package
// under. This is the count that really says how much of a package a program
// can reach.
func SupportedStdlibSymbols() map[string][]string {
	symbols := SupportedStdlibPackageFuncs()

	collect := func(store map[string]string) {
		for key := range store {
			dot := strings.LastIndex(key, ".")
			if dot <= 0 || dot == len(key)-1 {
				continue
			}

			pkg, name := key[:dot], key[dot+1:]
			list := symbols[pkg]
			found := false
			for _, existing := range list {
				if existing == name {
					found = true
					break
				}
			}
			if !found {
				symbols[pkg] = append(list, name)
			}
		}
	}

	collect(packageConstants)
	collect(packageTypes)
	collect(reportOnlyTypes)

	for pkg := range symbols {
		sort.Strings(symbols[pkg])
	}

	return symbols
}

// reportOnlyTypes are the named types the runtime answers for with a value of
// its own but that are not built by a registered constructor, such as
// strings.Reader, which is handed back by the NewReader function that makes it.
// They are counted as covered in the support report even though a zero value of
// them has no constructor of its own to be built from.
var reportOnlyTypes = map[string]string{
	"strings.Reader":   "",
	"strings.Replacer": "",
	"bytes.Reader":     "",
}

func isStdlibPkgAlias(name string) bool {
	for _, aliases := range stdlibPkgAliases {
		for _, alias := range aliases {
			if alias == name {
				return true
			}
		}
	}

	return false
}
