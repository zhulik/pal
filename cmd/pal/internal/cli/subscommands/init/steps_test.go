package initcmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	initcmd "github.com/zhulik/pal/cmd/pal/internal/cli/subscommands/init"
)

const (
	flagNoInteractive = "no-interactive"
	flagDirectory     = "directory"
	flagNoGit         = "no-git"
	flagTemplate      = "template"
	argModule         = "module"
	defaultTemplate   = "cli"
)

func TestStepsInputsCoveredByCLI(t *testing.T) {
	t.Parallel()

	flagNames := map[string]struct{}{}
	for _, f := range initcmd.Flags() {
		for _, name := range f.Names() {
			flagNames[name] = struct{}{}
		}
	}

	argNames := map[string]struct{}{}
	for _, a := range initcmd.Arguments() {
		argNames[a.(*cli.StringArg).Name] = struct{}{}
	}

	for _, step := range initcmd.Steps() {
		flag := step.Flag()
		arg := step.Argument()
		require.True(t, flag != nil || arg != nil, "step must expose a flag or argument")
		require.False(t, flag != nil && arg != nil, "step must not expose both a flag and an argument")

		if flag != nil {
			names := flag.Names()
			require.NotEmpty(t, names, "step flag must have a name")
			_, ok := flagNames[names[0]]
			require.True(t, ok, "step flag %q must appear in Flags()", names[0])
		}
		if arg != nil {
			stringArg, ok := arg.(*cli.StringArg)
			require.True(t, ok)
			_, ok = argNames[stringArg.Name]
			require.True(t, ok, "step argument %q must appear in Arguments()", stringArg.Name)
		}
	}
}

func TestFlagsIncludeNoInteractive(t *testing.T) {
	t.Parallel()

	var found bool
	for _, f := range initcmd.Flags() {
		for _, name := range f.Names() {
			if name == flagNoInteractive {
				found = true
			}
		}
	}
	require.True(t, found)

	// Mode flag is not a Step — ensure Steps do not own it.
	for _, step := range initcmd.Steps() {
		if flag := step.Flag(); flag != nil {
			for _, name := range flag.Names() {
				require.NotEqual(t, flagNoInteractive, name)
			}
		}
	}
}

func TestFlagsIncludeDirectory(t *testing.T) {
	t.Parallel()

	var found bool
	for _, f := range initcmd.Flags() {
		stringFlag, ok := f.(*cli.StringFlag)
		if !ok {
			continue
		}
		if stringFlag.Name == flagDirectory {
			found = true
			require.Equal(t, []string{"d"}, stringFlag.Aliases)
		}
	}
	require.True(t, found)

	for _, step := range initcmd.Steps() {
		if flag := step.Flag(); flag != nil {
			for _, name := range flag.Names() {
				require.NotEqual(t, flagDirectory, name)
				require.NotEqual(t, "d", name)
			}
		}
	}
}

func TestSteps_GitFlag(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagNoGit)
	boolFlag, ok := step.Flag().(*cli.BoolFlag)
	require.True(t, ok)
	require.Equal(t, flagNoGit, boolFlag.Name)
	require.Nil(t, step.Argument())
}

func TestSteps_ModuleArgument(t *testing.T) {
	t.Parallel()

	step := stepByArg(t, argModule)
	require.Nil(t, step.Flag())
	arg, ok := step.Argument().(*cli.StringArg)
	require.True(t, ok)
	require.Equal(t, argModule, arg.Name)
	require.Equal(t, "MODULE", arg.UsageText)
}

func TestSteps_ModuleApplicable(t *testing.T) {
	t.Parallel()

	step := stepByArg(t, argModule)
	dir := t.TempDir()
	require.True(t, step.Applicable(&initcmd.Options{}, dir))
	require.False(t, step.Applicable(&initcmd.Options{Module: "example.com/app"}, dir))
}

func TestSteps_GitApplicable(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagNoGit)
	dir := t.TempDir()
	require.True(t, step.Applicable(&initcmd.Options{}, dir))
	require.True(t, step.Applicable(&initcmd.Options{Git: false}, dir))
	require.False(t, step.Applicable(&initcmd.Options{GitSet: true, Git: false}, dir))
}

func TestSteps_GitDoesNotAbort(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagNoGit)
	require.False(t, step.Abort(&initcmd.Options{Git: false}))
	require.False(t, step.Abort(&initcmd.Options{Git: true}))
}

func TestSteps_ModuleDoesNotAbort(t *testing.T) {
	t.Parallel()

	step := stepByArg(t, argModule)
	require.False(t, step.Abort(&initcmd.Options{}))
	require.False(t, step.Abort(&initcmd.Options{Module: "example.com/app"}))
}

func TestSteps_TemplateFlag(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagTemplate)
	stringFlag, ok := step.Flag().(*cli.StringFlag)
	require.True(t, ok)
	require.Equal(t, flagTemplate, stringFlag.Name)
	require.Equal(t, defaultTemplate, stringFlag.Value)
	require.Nil(t, step.Argument())
}

func TestSteps_TemplateApplicable(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagTemplate)
	dir := t.TempDir()
	require.True(t, step.Applicable(&initcmd.Options{}, dir))
	require.True(t, step.Applicable(&initcmd.Options{Template: "cli"}, dir))
	require.False(t, step.Applicable(&initcmd.Options{TemplateSet: true, Template: "cli"}, dir))
}

func TestSteps_TemplateDoesNotAbort(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagTemplate)
	require.False(t, step.Abort(&initcmd.Options{}))
	require.False(t, step.Abort(&initcmd.Options{Template: "cli"}))
}

func TestSteps_TemplateFieldListsEmbeddedTemplates(t *testing.T) {
	t.Parallel()

	step := stepByFlag(t, flagTemplate)
	opts := &initcmd.Options{}
	field, err := step.Field(opts)
	require.NoError(t, err)
	require.NotNil(t, field)
	require.Equal(t, defaultTemplate, opts.Template)
}

func stepByFlag(t *testing.T, name string) initcmd.Step {
	t.Helper()
	for _, step := range initcmd.Steps() {
		if flag := step.Flag(); flag != nil {
			for _, n := range flag.Names() {
				if n == name {
					return step
				}
			}
		}
	}
	require.FailNowf(t, "step not found", "no step owns flag %q", name)
	return nil
}

func stepByArg(t *testing.T, name string) initcmd.Step {
	t.Helper()
	for _, step := range initcmd.Steps() {
		arg, ok := step.Argument().(*cli.StringArg)
		if ok && arg.Name == name {
			return step
		}
	}
	require.FailNowf(t, "step not found", "no step owns argument %q", name)
	return nil
}
