package javascript

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

func runtimeBundleSources() []string {
	return []string{
		runtimeSource(),
		collectionRuntimeSource(),
		rangeRuntimeSource(),
		genericRuntimeSource(),
		concurrencyRuntimeSource(),
		pathRuntimeSource(),
		bufioRuntimeSource(),
		randRuntimeSource(),
		cmpRuntimeSource(),
		errorsRuntimeSource(),
		extendedRuntimeSource(),
		extendedRuntimeSource2(),
		bytesToStringRuntimeSource(),
		osStdioRuntimeSource(),
		runeRuntimeSource(),
		moreRuntimeSource(),
	}
}

var runtimeFunctionPattern = regexp.MustCompile(`(?m)^function ([A-Za-z0-9_$]+)\(`)

func TestRuntimeSourcesHaveNoDuplicateFunctions(t *testing.T) {
	units := runtimeBundleSources()
	owners := map[string][]string{}

	for index, source := range units {
		for _, match := range runtimeFunctionPattern.FindAllStringSubmatch(source, -1) {
			owners[match[1]] = append(owners[match[1]], strconv.Itoa(index))
		}
	}

	var duplicates []string

	for name, seen := range owners {
		if len(seen) > 1 {
			duplicates = append(duplicates, name+" declared in units "+strings.Join(seen, ", "))
		}
	}

	sort.Strings(duplicates)

	for _, duplicate := range duplicates {
		t.Error("duplicate runtime function: " + duplicate)
	}
}

func TestRuntimeSourcesAreValidJavaScript(t *testing.T) {
	dir := t.TempDir()

	for index, source := range runtimeBundleSources() {
		path := filepath.Join(dir, fmt.Sprintf("unit_%02d.js", index))

		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}

		check := exec.Command("node", "--check", path)

		if output, err := check.CombinedOutput(); err != nil {
			t.Errorf("runtime unit %d is not valid JavaScript: %v\n%s", index, err, output)
		}
	}
}

func TestPackageVarValuesHaveMatchingTypes(t *testing.T) {
	moreStdlibFuncs()

	for key := range packageVarValues {
		if packageVarTypes[key] == "" {
			t.Errorf("package var %s has no packageVarTypes entry", key)
		}
	}

	for key := range packageVarTypes {
		if packageVarValues[key] == "" {
			t.Errorf("package var type %s has no packageVarValues entry", key)
		}
	}
}

func TestPackageVarValuesEmitSingleCall(t *testing.T) {
	moreStdlibFuncs()

	for key, value := range packageVarValues {
		if strings.HasSuffix(value, "()") {
			t.Errorf("package var %s maps to %q, which would emit a doubled call", key, value)
		}
	}
}
