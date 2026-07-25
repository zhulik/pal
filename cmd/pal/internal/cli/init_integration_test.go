package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal/cmd/pal/templates"
)

const integrationModule = "example.com/paltest"

// templateValidators maps each scaffold name to a post-init mutator/validator.
// Adding a template under scaffolds/ requires a matching entry here.
var templateValidators = map[string]func(t *testing.T, dir, pkg string){
	"cli": validateCLITemplate,
}

func TestInit_Integration(t *testing.T) {
	t.Parallel()

	names, err := templates.Names()
	require.NoError(t, err)
	require.NotEmpty(t, names)

	root := repoRoot(t)
	bin := buildPalBinary(t, root)
	pkg := templates.PackageName(integrationModule)

	type dirMode struct {
		name string
		make func(t *testing.T) string
	}
	dirModes := []dirMode{
		{
			name: "existing-empty",
			make: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
		},
		{
			name: "missing-nested",
			make: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "new", "nested", "app")
			},
		},
	}

	for _, template := range names {
		validator, ok := templateValidators[template]
		require.Truef(t, ok, "no integration validator for template %q — add one to templateValidators", template)

		for _, git := range []bool{true, false} {
			for _, mode := range dirModes {
				t.Run(template+"/git="+boolStr(git)+"/"+mode.name, func(t *testing.T) {
					t.Parallel()

					dir := mode.make(t)
					runPalInit(t, bin, dir, template, git)
					assertGitState(t, dir, git)
					replaceLocalPal(t, dir, root)
					validator(t, dir, pkg)
					requireProjectTooling(t, dir, root)
				})
			}
		}
	}
}

func runPalInit(t *testing.T, bin, dir, template string, git bool) {
	t.Helper()

	args := []string{
		"init",
		"--no-interactive",
		"-d", dir,
		integrationModule,
		"--template", template,
	}
	if !git {
		args = append(args, "--no-git")
	}

	runCmd(t, "", bin, args...)
}

func assertGitState(t *testing.T, dir string, wantGit bool) {
	t.Helper()
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if wantGit {
		require.NoError(t, err)
		require.True(t, info.IsDir())
		return
	}
	require.ErrorIs(t, err, os.ErrNotExist)
}

func replaceLocalPal(t *testing.T, dir, root string) {
	t.Helper()
	runCmd(t, dir, "go", "mod", "edit", "-replace", "github.com/zhulik/pal="+root)
	runCmd(t, dir, "go", "mod", "tidy")
}

func validateCLITemplate(t *testing.T, dir, pkg string) {
	t.Helper()

	mainPath := filepath.Join(dir, "cmd", pkg, "main.go")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	old := `"` + pkg + ` running"`
	newMsg := `"` + pkg + ` integration-ok"`
	require.Contains(t, string(data), old)
	updated := strings.Replace(string(data), old, newMsg, 1)
	require.NoError(t, os.WriteFile(mainPath, []byte(updated), 0o644))

	testPath := filepath.Join(dir, "cmd", pkg, "main_test.go")
	smoke := `package main_test

import "testing"

func TestSmoke(t *testing.T) {
	t.Parallel()
	if got := "` + pkg + `"; got == "" {
		t.Fatal("empty package name")
	}
}
`
	require.NoError(t, os.WriteFile(testPath, []byte(smoke), 0o644))
}

func requireProjectTooling(t *testing.T, dir, root string) {
	t.Helper()
	// Keep asdf (and similar) on the same pins as the repo when cwd is a temp tree.
	if data, err := os.ReadFile(filepath.Join(root, ".tool-versions")); err == nil {
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".tool-versions"), data, 0o644))
	}
	// golangci-lint refuses concurrent runs; serialize task across parallel cases.
	projectToolingMu.Lock()
	defer projectToolingMu.Unlock()
	runCmd(t, dir, "task")
}

var projectToolingMu sync.Mutex

func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	out, err := cmdOutput(t, dir, name, args...)
	require.NoErrorf(t, err, "%s %s\n%s", name, strings.Join(args, " "), out)
}

func cmdOutput(t *testing.T, dir, name string, args ...string) ([]byte, error) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		mod := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(mod)
		if err == nil && strings.HasPrefix(string(data), "module github.com/zhulik/pal\n") {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "repo root with module github.com/zhulik/pal not found from %s", dir)
		dir = parent
	}
}

var (
	palBinOnce sync.Once
	palBinPath string
	palBinErr  error
	palBinOut  []byte
)

func buildPalBinary(t *testing.T, root string) string {
	t.Helper()
	palBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "pal-cli-integration-bin-*")
		if err != nil {
			palBinErr = err
			return
		}
		palBinPath = filepath.Join(dir, "pal")
		cmd := exec.Command("go", "build", "-o", palBinPath, ".")
		cmd.Dir = filepath.Join(root, "cmd", "pal")
		palBinOut, palBinErr = cmd.CombinedOutput()
	})
	require.NoErrorf(t, palBinErr, "go build pal\n%s", palBinOut)
	return palBinPath
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
