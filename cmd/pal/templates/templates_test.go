package templates_test

import (
	"os"
	"path/filepath"
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

func TestApplyCLI(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := templates.Apply(dir, "cli", templates.Data{Package: "myapp"})
	require.NoError(t, err)

	mainPath := filepath.Join(dir, "cmd", "myapp", "main.go")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "package main")
	require.Contains(t, string(data), `"myapp running"`)
	require.NotContains(t, string(data), "{{.Package}}")
}

func TestApplyUnknown(t *testing.T) {
	t.Parallel()

	err := templates.Apply(t.TempDir(), "missing", templates.Data{Package: "x"})
	require.Error(t, err)
}
