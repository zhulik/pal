package templates

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:scaffolds
var scaffoldsFS embed.FS

// FS is the project template tree (one directory per template name under
// scaffolds/). Adding a template is: create scaffolds/<name>/… — no new
// embed directive is required.
var FS = mustSub(scaffoldsFS, "scaffolds")

func mustSub(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

// Data is the values available to project template paths and file contents.
type Data struct {
	// Package is the Go package / binary name derived from the module path.
	Package string
	// Module is the full Go module path (e.g. github.com/user/app).
	Module string
}

// Apply renders the named project template from FS into dst.
// Both relative paths and file contents are Go text/template templates
// executed with data. Every scaffold file must use a trailing ".tmpl"
// suffix; Apply strips it from the output path (so Go sources are ignored by
// go test/list, and all scaffold files share one convention).
func Apply(dst, name string, data Data) error {
	if name == "" {
		return fmt.Errorf("template name is required")
	}
	if _, err := fs.Stat(FS, name); err != nil {
		return fmt.Errorf("unknown template %q: %w", name, err)
	}

	return fs.WalkDir(FS, name, func(embedPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(name, embedPath)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		renderedRel, err := execute(filepath.ToSlash(rel), data)
		if err != nil {
			return fmt.Errorf("render path %q: %w", rel, err)
		}
		if !d.IsDir() {
			renderedRel = strings.TrimSuffix(renderedRel, ".tmpl")
		}
		outPath := filepath.Join(dst, filepath.FromSlash(renderedRel))

		if d.IsDir() {
			return os.MkdirAll(outPath, 0o755)
		}

		raw, err := fs.ReadFile(FS, embedPath)
		if err != nil {
			return err
		}
		rendered, err := execute(string(raw), data)
		if err != nil {
			return fmt.Errorf("render file %q: %w", embedPath, err)
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		//nolint:gosec // G306: scaffolded project sources are intentionally world-readable
		return os.WriteFile(outPath, []byte(rendered), 0o644)
	})
}

// Names returns the project template directory names embedded in FS.
func Names() ([]string, error) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// PackageName derives the Go package name from a module path (last element,
// with '-' replaced by '_').
func PackageName(module string) string {
	return strings.ReplaceAll(path.Base(module), "-", "_")
}

func execute(text string, data Data) (string, error) {
	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
