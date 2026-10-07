package main

import (
	"testing"

	"github.com/zenobiatranoss/go2js/backend/javascript"
)

// TestExportDataLoadable pins the heart of the report: the installed Go's
// export data for a supported package must load, so coverage numbers are real.
func TestExportDataLoadable(t *testing.T) {
	for _, path := range []string{"fmt", "strings", "strconv", "math", "time", "os"} {
		names := exportedNames(path)
		if len(names) == 0 {
			t.Errorf("exportedNames(%q) empty; export data did not load", path)
		}
	}
}

// TestBucketHead guards the boundary between the runtime registry and
// this tool: every inventory name must resolve to some import path, the
// identifier keys resolve to their import path, and keys that are already
// import paths resolve to themselves.
func TestImportPathFor(t *testing.T) {
	inventory := javascript.SupportedStdlibPackageFuncs()
	for name := range inventory {
		path := importPathFor(name, inventory)
		if path == "" {
			t.Errorf("importPathFor(%q) resolved to empty path", name)
		}
	}

	for ident, path := range identPath {
		if importPathFor(ident, inventory) != path {
			t.Errorf("importPathFor(%q) != %q", ident, path)
		}
	}

	for _, path := range []string{"fmt", "strconv", "encoding/base64", "unicode/utf16"} {
		if got := importPathFor(path, inventory); got != path {
			t.Errorf("importPathFor(%q) != itself (%q)", path, got)
		}
	}
}
