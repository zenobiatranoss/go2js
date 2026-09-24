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

func moduleExport(name string, format ModuleFormat) string {
	if name == "" {
		return ""
	}

	switch format {
	case ModuleCommonJS:
		return fmt.Sprintf("module.exports.%s = %s;\n", name, name)
	case ModuleIIFE:
		return ""
	default:
		return fmt.Sprintf("export { %s };\n", name)
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

func Export(name string, options ModuleOptions) string {
	options = options.Normalize()
	return moduleExport(name, options.Format)
}

func Import(name, path string, options ModuleOptions) string {
	options = options.Normalize()
	return moduleImport(name, path, options.Format)
}
