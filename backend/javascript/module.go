package javascript

import (
	"fmt"
	"strings"
)

type ModuleFormat string

const (
	ModuleESM      ModuleFormat = "esm"
	ModuleCommonJS ModuleFormat = "commonjs"
	ModuleIIFE     ModuleFormat = "iife"
)

type ModuleOptions struct {
	Format     ModuleFormat
	Strict     bool
	ExportMain bool
	Name       string
}

func DefaultModuleOptions() ModuleOptions {
	return ModuleOptions{
		Format: ModuleESM,
		Strict: true,
	}
}

func NewModuleOptions(format string) ModuleOptions {
	return ModuleOptions{
		Format: normalizeModuleFormat(format),
		Strict: true,
	}
}

func (o ModuleOptions) Normalize() ModuleOptions {
	o.Format = normalizeModuleFormat(string(o.Format))
	return o
}

func (o ModuleOptions) IsESM() bool {
	return o.Normalize().Format == ModuleESM
}

func (o ModuleOptions) IsCommonJS() bool {
	return o.Normalize().Format == ModuleCommonJS
}

func (o ModuleOptions) IsIIFE() bool {
	return o.Normalize().Format == ModuleIIFE
}

func normalizeModuleFormat(format string) ModuleFormat {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "commonjs", "common-js", "cjs":
		return ModuleCommonJS
	case "iife":
		return ModuleIIFE
	default:
		return ModuleESM
	}
}

func moduleHeader(options ModuleOptions) string {
	if !options.Strict {
		return ""
	}

	return "\"use strict\";\n"
}

func modulePrefix(options ModuleOptions) string {
	if options.Format == ModuleIIFE {
		return "(function() {\n"
	}

	return ""
}

func moduleFooter(options ModuleOptions) string {
	if options.Format == ModuleIIFE {
		return "})();\n"
	}

	return ""
}

// moduleExportAs is one declaration handed over under a name of its own, which is
// the name it is declared under inside the module and the name it is reached
// under outside it, which are two names when a declaration was written asking for
// them to be.
func moduleExportAs(name, alias string, format ModuleFormat) string {
	if name == "" {
		return ""
	}

	exported := alias

	if exported == "" {
		exported = name
	}

	switch format {
	case ModuleCommonJS:
		return fmt.Sprintf("module.exports.%s = %s;\n", exported, name)
	case ModuleIIFE:
		return ""
	default:
		if exported == name {
			return fmt.Sprintf("export { %s };\n", name)
		}

		return fmt.Sprintf("export { %s as %s };\n", name, exported)
	}
}

func moduleImport(name, path string, format ModuleFormat) string {
	if name == "" || path == "" {
		return ""
	}

	switch format {
	case ModuleCommonJS:
		return fmt.Sprintf("const %s = require(%q);\n", name, path)
	case ModuleIIFE:
		return ""
	default:
		return fmt.Sprintf("import %s from %q;\n", name, path)
	}
}

func WrapModule(source string, options ModuleOptions) string {
	options = options.Normalize()

	var b strings.Builder
	b.WriteString(moduleHeader(options))
	b.WriteString(modulePrefix(options))
	b.WriteString(source)

	if source != "" && !strings.HasSuffix(source, "\n") {
		b.WriteByte('\n')
	}

	b.WriteString(moduleFooter(options))
	return b.String()
}

func Import(name, path string, options ModuleOptions) string {
	options = options.Normalize()
	return moduleImport(name, path, options.Format)
}
