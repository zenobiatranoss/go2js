// Package bindings holds the bindings that go2js ships: libraries a program can
// use with -with NAME, without generating anything. Each was written by
// go2js bind (scripts/gen-bindings.sh says how), so a different cut of one can be
// made the same way.
package bindings

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

//go:embed dom/dom.stub dom/dom.js
var files embed.FS

// Names is the bindings there are.
func Names() []string {
	entries, err := files.ReadDir(".")

	if err != nil {
		return nil
	}

	var names []string

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	sort.Strings(names)

	return names
}

// Install writes the stub and the shim of a binding into dir and gives the
// path of the shim.
func Install(name, dir string) (shim string, err error) {
	stub, err := files.ReadFile(name + "/" + name + ".stub")

	if err != nil {
		return "", fmt.Errorf("go2js: there is no binding named %q (there are %v)", name, Names())
	}

	script, err := files.ReadFile(name + "/" + name + ".js")

	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(dir, name+".stub"), stub, 0o644); err != nil {
		return "", err
	}

	shim = filepath.Join(dir, name+".js")

	return shim, os.WriteFile(shim, script, 0o644)
}
