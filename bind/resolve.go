package bind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reference = regexp.MustCompile(`(?m)(?:^\s*(?:export|import)\b[^;'"]*?\bfrom\s*|^\s*import\s*|///\s*<reference\s+path\s*=\s*)["'](\.[^"']*)["']`)

// FindDeclarations finds the declaration files of an npm package that is
// installed under dir: the file its package.json names as its types, or the
// package's own index.d.ts, or the one the DefinitelyTyped project wrote for it
// (@types/name). The files it points to by relative paths are found with it.
func FindDeclarations(dir, name string) ([]string, error) {
	var roots []string

	for _, base := range ancestors(dir) {
		modules := filepath.Join(base, "node_modules")

		if entry := typesOf(filepath.Join(modules, filepath.FromSlash(name))); entry != "" {
			roots = append(roots, entry)

			break
		}

		scoped := strings.TrimPrefix(strings.Replace(name, "/", "__", 1), "@")

		if entry := typesOf(filepath.Join(modules, "@types", scoped)); entry != "" {
			roots = append(roots, entry)

			break
		}
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("bind: no declaration files for %q under %s: install the package (npm install %s), or give its .d.ts file with -dts", name, filepath.Join(dir, "node_modules"), name)
	}

	return collect(roots), nil
}

func ancestors(dir string) []string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}

	var out []string

	for {
		out = append(out, abs)

		parent := filepath.Dir(abs)

		if parent == abs {
			return out
		}

		abs = parent
	}
}

// typesOf is the declaration file a package directory has, or "".
func typesOf(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}

	var manifest struct {
		Types   string `json:"types"`
		Typings string `json:"typings"`
		Main    string `json:"main"`
	}

	_ = json.Unmarshal(data, &manifest)

	for _, entry := range []string{manifest.Types, manifest.Typings} {
		if entry == "" {
			continue
		}

		if path := declarationFile(filepath.Join(dir, filepath.FromSlash(entry))); path != "" {
			return path
		}
	}

	if manifest.Main != "" {
		if path := declarationFile(strings.TrimSuffix(filepath.Join(dir, filepath.FromSlash(manifest.Main)), ".js")); path != "" {
			return path
		}
	}

	return declarationFile(filepath.Join(dir, "index"))
}

// declarationFile is the file a path stands for as a module of declarations.
func declarationFile(path string) string {
	for _, candidate := range []string{path, path + ".d.ts", filepath.Join(path, "index.d.ts")} {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() && strings.HasSuffix(candidate, ".d.ts") {
			return candidate
		}
	}

	return ""
}

// collect adds the files that the given ones refer to by a relative path.
func collect(roots []string) []string {
	var order []string

	seen := map[string]bool{}
	queue := append([]string(nil), roots...)

	for len(queue) > 0 && len(order) < 300 {
		path := queue[0]
		queue = queue[1:]

		if seen[path] {
			continue
		}

		seen[path] = true
		order = append(order, path)

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, match := range reference.FindAllStringSubmatch(string(data), -1) {
			target := filepath.Join(filepath.Dir(path), filepath.FromSlash(match[1]))

			if found := declarationFile(strings.TrimSuffix(target, ".js")); found != "" {
				queue = append(queue, found)
			}
		}
	}

	return order
}
