package javascript

import (
	"strings"
	"testing"
)

func TestRuntimeBundleSelectsRequiredSourceUnits(t *testing.T) {
	source := runtimeBundle(
		`function main(){go2jsLen(values);}`,
		`function go2jsLen(value) {
	return value.length;
}

function go2jsCap(value) {
	return value.length;
}`,
		`const go2jsUnusedState = new Map();

function go2jsOther(value) {
	return value;
}`,
	)

	if !strings.Contains(source, "function go2jsLen") {
		t.Fatal("required runtime helper missing")
	}

	if strings.Contains(source, "function go2jsOther") {
		t.Fatal("unused runtime source unit was included")
	}
}

func TestRuntimeBundleIncludesDependenciesAndGlobals(t *testing.T) {
	source := runtimeBundle(
		`function main(){go2jsSliceCap(value);}`,
		`function go2jsPtr(value) {
	return value;
}`,
		`const go2jsSliceMeta = new WeakMap();

function go2jsSliceState(value) {
	return go2jsSliceMeta.get(value);
}

function go2jsSliceCap(value) {
	return go2jsSliceState(value).capacity;
}`,
	)

	if !strings.Contains(source, "const go2jsSliceMeta = new WeakMap()") {
		t.Fatal("required runtime global missing")
	}

	if !strings.Contains(source, "function go2jsSliceState") {
		t.Fatal("runtime dependency missing")
	}

	if !strings.Contains(source, "function go2jsSliceCap") {
		t.Fatal("root runtime helper missing")
	}

	if strings.Contains(source, "function go2jsPtr") {
		t.Fatal("unused runtime source unit was included")
	}
}

func TestRuntimeBundleIncludesGeneratorFunctions(t *testing.T) {
	source := runtimeBundle(
		`function main(){for(const item of go2jsRangeValue(value)){};}`,
		`function* go2jsRangeValue(value) {
		yield [0, value];
	}`,
	)

	if !strings.Contains(source, "function* go2jsRangeValue") {
		t.Fatal("generator runtime helper missing")
	}
}

func TestRuntimeBundlePreservesSourceOrder(t *testing.T) {
	source := runtimeBundle(
		`go2jsSecond();`,
		`function go2jsFirst() {
	return 1;
}`,
		`function go2jsSecond() {
	return go2jsFirst();
}`,
	)

	first := strings.Index(source, "function go2jsFirst")
	second := strings.Index(source, "function go2jsSecond")

	if first < 0 || second < 0 {
		t.Fatal("expected runtime helpers")
	}

	if first > second {
		t.Fatal("runtime source order changed")
	}
}
