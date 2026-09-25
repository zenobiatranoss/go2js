package javascript

import (
	"regexp"
	"strings"
)

var runtimeFunctionHeader = regexp.MustCompile(`(?m)^\s*function\*?\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\([^)]*\)\s*\{`)
var runtimeDependency = regexp.MustCompile(`\bgo2js[A-Za-z0-9_$]*\b`)

type runtimeSourceUnit struct {
	source    string
	functions map[string]bool
}

func runtimeBundle(requiredSource string, sources ...string) string {
	units := make([]runtimeSourceUnit, 0, len(sources))

	for _, source := range sources {
		functions := make(map[string]bool)

		for _, match := range runtimeFunctionHeader.FindAllStringSubmatch(source, -1) {
			functions[match[1]] = true
		}

		units = append(units, runtimeSourceUnit{
			source:    source,
			functions: functions,
		})
	}

	required := make(map[string]bool)

	for _, name := range runtimeDependency.FindAllString(requiredSource, -1) {
		required[name] = true
	}

	included := make([]bool, len(units))
	changed := true

	for changed {
		changed = false

		for i, unit := range units {
			if included[i] {
				continue
			}

			needed := false

			for name := range required {
				if unit.functions[name] {
					needed = true
					break
				}
			}

			if !needed {
				continue
			}

			included[i] = true
			changed = true

			for _, name := range runtimeDependency.FindAllString(unit.source, -1) {
				required[name] = true
			}
		}
	}

	var out strings.Builder

	for i, unit := range units {
		if !included[i] {
			continue
		}

		if out.Len() > 0 {
			out.WriteString("\n\n")
		}

		out.WriteString(strings.TrimSpace(unit.source))
	}

	return out.String()
}
