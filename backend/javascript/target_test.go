package javascript

import (
	"strings"
	"testing"
)

func TestLowerJavaScriptTargetES2019(t *testing.T) {
	source := `return value.constructor?.name || "object";
const previousCap = value.__go2js_cap ?? value.length;`

	output := lowerJavaScriptTarget(source, "es2019")

	if strings.Contains(output, "?.") {
		t.Fatal("ES2019 output still contains optional chaining")
	}

	if strings.Contains(output, "??") {
		t.Fatal("ES2019 output still contains nullish coalescing")
	}

	if !strings.Contains(output, "value.constructor == null") {
		t.Fatal("optional chaining was not lowered")
	}

	if !strings.Contains(output, "value.__go2js_cap != null") {
		t.Fatal("nullish coalescing was not lowered")
	}
}

func TestLowerJavaScriptTargetModern(t *testing.T) {
	source := `return value.constructor?.name || "object";
const previousCap = value.__go2js_cap ?? value.length;`

	for _, target := range []string{
		"es2020",
		"es2021",
		"es2022",
		"es2023",
		"es2024",
		"esnext",
	} {
		output := lowerJavaScriptTarget(source, target)

		if output != source {
			t.Fatalf("target %s unexpectedly lowered modern syntax", target)
		}
	}
}

func TestEmitWithContextOptionsTargetDefaults(t *testing.T) {
	source := `function main() {
    return value.constructor?.name || "object";
}`

	output := lowerJavaScriptTarget(source, "es2019")

	if strings.Contains(output, "?.") {
		t.Fatal("default target lowering failed")
	}
}
