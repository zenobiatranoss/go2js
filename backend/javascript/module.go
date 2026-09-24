package javascript

import (
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
}

func DefaultModuleOptions() ModuleOptions {
	return ModuleOptions{
		Format: ModuleESM,
		Strict: true,
	}
}

func normalizeModuleFormat(format string) ModuleFormat {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "commonjs", "cjs":
		return ModuleCommonJS
	case "iife":
		return ModuleIIFE
	default:
		return ModuleESM
	}
}

func moduleHeader(options ModuleOptions) string {
	var b strings.Builder
	if options.Strict {
		b.WriteString("\"use strict\";\n")
	}
	return b.String()
}

func moduleExport(name string, format ModuleFormat) string {
	switch format {
	case ModuleCommonJS:
		return "module.exports." + name + " = " + name + ";\n"
	case ModuleIIFE:
		return ""
	default:
		return "export { " + name + " };\n"
	}
}
