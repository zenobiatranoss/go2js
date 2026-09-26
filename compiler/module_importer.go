package compiler

import (
	"encoding/json"
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"sync"
)

// moduleImporter resolves non-local imports by asking the go tool for the
// compiled export data of the package inside the active module. The default
// importer only searches GOROOT and GOPATH, so third-party modules living in
// the module cache are invisible without this.
type moduleImporter struct {
	root  string
	inner types.Importer

	mu    sync.Mutex
	cache map[string]*types.Package
}

func newModuleImporter(root string) *moduleImporter {
	fileSet := token.NewFileSet()

	m := &moduleImporter{
		root:  root,
		cache: map[string]*types.Package{},
	}
	m.inner = importer.ForCompiler(fileSet, "gc", m.lookupExportData)

	return m
}

type goListPackage struct {
	Export string
}

func (m *moduleImporter) lookupExportData(path string) (io.ReadCloser, error) {
	archive, err := m.exportArchive(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(archive)
	if err != nil {
		return nil, fmt.Errorf("compiler: reading export data for %q: %w", path, err)
	}

	return file, nil
}

func (m *moduleImporter) exportArchive(path string) (string, error) {
	cmd := exec.Command("go", "list", "-export", "-json", path)
	cmd.Dir = m.root

	stdout, err := cmd.Output()
	if err != nil {
		combined, _ := cmd.CombinedOutput()

		message := goListErrorMessage(combined)
		if message == "" {
			message = err.Error()
		}

		return "", fmt.Errorf("compiler: locating package %q: %s", path, message)
	}

	var result goListPackage
	if err := json.Unmarshal(stdout, &result); err != nil {
		return "", fmt.Errorf("compiler: decoding go list output for %q: %w", path, err)
	}

	if result.Export == "" {
		return "", fmt.Errorf("compiler: package %q produced no export data", path)
	}

	return result.Export, nil
}

func (m *moduleImporter) Import(path string) (*types.Package, error) {
	if path == "" {
		return nil, fmt.Errorf("compiler: empty import path")
	}

	m.mu.Lock()
	if pkg, ok := m.cache[path]; ok {
		m.mu.Unlock()
		return pkg, nil
	}
	m.mu.Unlock()

	pkg, err := m.inner.Import(path)
	if err != nil {
		return nil, fmt.Errorf("compiler: importing %q: %w", path, err)
	}

	m.mu.Lock()
	m.cache[path] = pkg
	m.mu.Unlock()

	return pkg, nil
}

func goListErrorMessage(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var entries []struct {
		Err string `json:"Err"`
	}

	if err := json.Unmarshal(data, &entries); err == nil {
		for _, entry := range entries {
			if entry.Err != "" {
				return entry.Err
			}
		}
	}

	return string(data)
}
