package javascript

import "strings"

func normalizeTarget(target string) string {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return "es2022"
	}
	return target
}

func lowerJavaScriptTarget(source, target string) string {
	target = normalizeTarget(target)

	if target != "es2019" {
		return source
	}

	source = strings.ReplaceAll(
		source,
		`return value.constructor?.name || "object";`,
		`return (value.constructor == null ? undefined : value.constructor.name) || "object";`,
	)

	source = strings.ReplaceAll(
		source,
		`const previousCap = value.__go2js_cap ?? value.length;`,
		`const previousCap = value.__go2js_cap != null ? value.__go2js_cap : value.length;`,
	)

	return source
}
