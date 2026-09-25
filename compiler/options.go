package compiler

import (
	"fmt"
	"strings"
)

type Options struct {
	PackageName string
	Module      string
	SourceMap   bool
	Minify      bool
	Strict      bool
	Pretty      bool
	Runtime     bool
	Target      string
}

func DefaultOptions() Options {
	return Options{
		Module:  "esm",
		Strict:  true,
		Pretty:  true,
		Runtime: true,
		Target:  "es2022",
	}
}

func (o Options) WithPackageName(name string) Options {
	o.PackageName = strings.TrimSpace(name)
	return o
}

func (o Options) WithModule(module string) Options {
	o.Module = strings.ToLower(strings.TrimSpace(module))
	return o
}

func (o Options) WithSourceMap(enabled bool) Options {
	o.SourceMap = enabled
	return o
}

func (o Options) WithMinify(enabled bool) Options {
	o.Minify = enabled
	if enabled {
		o.Pretty = false
	}
	return o
}

func (o Options) WithStrict(enabled bool) Options {
	o.Strict = enabled
	return o
}

func (o Options) WithPretty(enabled bool) Options {
	o.Pretty = enabled
	if enabled {
		o.Minify = false
	}
	return o
}

func (o Options) WithRuntime(enabled bool) Options {
	o.Runtime = enabled
	return o
}

func (o Options) WithTarget(target string) Options {
	o.Target = strings.ToLower(strings.TrimSpace(target))
	return o
}

func (o Options) Normalize() Options {
	if strings.TrimSpace(o.Module) == "" {
		o.Module = "esm"
	}
	if strings.TrimSpace(o.Target) == "" {
		o.Target = "es2022"
	}

	o.Module = normalizeModule(o.Module)
	o.Target = strings.ToLower(strings.TrimSpace(o.Target))

	if o.Minify {
		o.Pretty = false
	}
	if !o.Minify && !o.Pretty {
		o.Pretty = true
	}

	return o
}

func (o Options) Validate() error {
	o = o.Normalize()

	switch o.Module {
	case "esm", "commonjs", "iife":
	default:
		return fmt.Errorf("unsupported module format %q", o.Module)
	}

	switch o.Target {
	case "es2019", "es2020", "es2021", "es2022", "es2023", "es2024", "esnext":
	default:
		return fmt.Errorf("unsupported target %q", o.Target)
	}

	if o.PackageName != "" {
		for _, r := range o.PackageName {
			if !(r == '_' || r == '-' || r == '.' ||
				r >= 'a' && r <= 'z' ||
				r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9') {
				return fmt.Errorf("invalid package name %q", o.PackageName)
			}
		}
	}

	return nil
}

func normalizeModule(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	switch value {
	case "common-js", "commonjs", "cjs":
		return "commonjs"
	case "module", "modules", "esm":
		return "esm"
	case "self", "iife":
		return "iife"
	default:
		return value
	}
}

func (o Options) IsESM() bool {
	return o.Normalize().Module == "esm"
}

func (o Options) IsCommonJS() bool {
	return o.Normalize().Module == "commonjs"
}

func (o Options) IsIIFE() bool {
	return o.Normalize().Module == "iife"
}

func (o Options) TargetVersion() string {
	return o.Normalize().Target
}
