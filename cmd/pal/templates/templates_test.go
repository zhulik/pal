package templates_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zhulik/pal/cmd/pal/templates"
)

func TestNames(t *testing.T) {
	t.Parallel()

	names, err := templates.Names()
	require.NoError(t, err)
	require.Equal(t, []string{"cli"}, names)
}

func TestPackageName(t *testing.T) {
	t.Parallel()

	require.Equal(t, "myapp", templates.PackageName("example.com/myapp"))
	require.Equal(t, "my_app", templates.PackageName("example.com/my-app"))
}

func TestScaffoldFilesHaveTmplSuffix(t *testing.T) {
	t.Parallel()

	err := fs.WalkDir(templates.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		require.True(t, strings.HasSuffix(path, ".tmpl"), "scaffold file %q must end with .tmpl", path)
		return nil
	})
	require.NoError(t, err)
}

func TestApplyCLI(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := templates.Apply(dir, "cli", templates.Data{
		Package: "myapp",
		Module:  "example.com/myapp",
	})
	require.NoError(t, err)

	mainPath := filepath.Join(dir, "cmd", "myapp", "main.go")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "package main")
	require.Contains(t, string(data), `"myapp running"`)
	require.NotContains(t, string(data), "{{.Package}}")

	for _, rel := range []string{
		"README.md",
		".gitignore",
		"Taskfile.yaml",
		".golangci.yaml",
	} {
		path := filepath.Join(dir, rel)
		_, err := os.Stat(path)
		require.NoError(t, err, "expected scaffold file %s", rel)
		require.False(t, strings.HasSuffix(path, ".tmpl"))
	}

	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(readme), "# myapp")
	require.Contains(t, string(readme), "example.com/myapp")
	require.Contains(t, string(readme), "go run ./cmd/myapp")
	require.NotContains(t, string(readme), "{{.")

	taskfile, err := os.ReadFile(filepath.Join(dir, "Taskfile.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(taskfile), "go run ./cmd/myapp")
	require.NotContains(t, string(taskfile), "{{.")
}

func TestApplyUnknown(t *testing.T) {
	t.Parallel()

	err := templates.Apply(t.TempDir(), "missing", templates.Data{Package: "x"})
	require.Error(t, err)
}

func TestExists(t *testing.T) {
	t.Parallel()

	require.NoError(t, templates.Exists("cli"))
	require.Error(t, templates.Exists(""))
	require.Error(t, templates.Exists("missing"))
}
