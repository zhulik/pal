package initcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestStepsInputsCoveredByCLI(t *testing.T) {
	t.Parallel()

	flagNames := map[string]struct{}{}
	for _, f := range Flags() {
		for _, name := range f.Names() {
			flagNames[name] = struct{}{}
		}
	}

	argNames := map[string]struct{}{}
	for _, a := range Arguments() {
		argNames[a.(*cli.StringArg).Name] = struct{}{}
	}

	for _, step := range Steps() {
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
	for _, f := range Flags() {
		for _, name := range f.Names() {
			if name == flagNoInteractive {
				found = true
			}
		}
	}
	require.True(t, found)

	// Mode flag is not a Step — ensure Steps do not own it.
	for _, step := range Steps() {
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
	for _, f := range Flags() {
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

	for _, step := range Steps() {
		if flag := step.Flag(); flag != nil {
			for _, name := range flag.Names() {
				require.NotEqual(t, flagDirectory, name)
				require.NotEqual(t, "d", name)
			}
		}
	}
}

func TestGitStepFlag(t *testing.T) {
	t.Parallel()

	flag := gitStep{}.Flag()
	boolFlag, ok := flag.(*cli.BoolFlag)
	require.True(t, ok)
	require.Equal(t, flagNoGit, boolFlag.Name)
	require.Nil(t, gitStep{}.Argument())
}

func TestModuleStepArgument(t *testing.T) {
	t.Parallel()

	require.Nil(t, moduleStep{}.Flag())
	arg, ok := moduleStep{}.Argument().(*cli.StringArg)
	require.True(t, ok)
	require.Equal(t, argModule, arg.Name)
	require.Equal(t, "MODULE", arg.UsageText)
}

func TestModuleStepApplicable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.True(t, moduleStep{}.Applicable(&Options{}, dir))
	require.False(t, moduleStep{}.Applicable(&Options{Module: "example.com/app"}, dir))
}

func TestGitStepApplicable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.True(t, gitStep{}.Applicable(&Options{}, dir))
	require.True(t, gitStep{}.Applicable(&Options{Git: false}, dir))
	require.False(t, gitStep{}.Applicable(&Options{GitSet: true, Git: false}, dir))
}

func TestGitStepDoesNotAbort(t *testing.T) {
	t.Parallel()

	require.False(t, gitStep{}.Abort(&Options{Git: false}))
	require.False(t, gitStep{}.Abort(&Options{Git: true}))
}

func TestModuleStepDoesNotAbort(t *testing.T) {
	t.Parallel()

	require.False(t, moduleStep{}.Abort(&Options{}))
	require.False(t, moduleStep{}.Abort(&Options{Module: "example.com/app"}))
}

func TestTemplateStepFlag(t *testing.T) {
	t.Parallel()

	flag := templateStep{}.Flag()
	stringFlag, ok := flag.(*cli.StringFlag)
	require.True(t, ok)
	require.Equal(t, flagTemplate, stringFlag.Name)
	require.Equal(t, defaultTemplate, stringFlag.Value)
	require.Nil(t, templateStep{}.Argument())
}

func TestTemplateStepApplicable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.True(t, templateStep{}.Applicable(&Options{}, dir))
	require.True(t, templateStep{}.Applicable(&Options{Template: "cli"}, dir))
	require.False(t, templateStep{}.Applicable(&Options{TemplateSet: true, Template: "cli"}, dir))
}

func TestTemplateStepDoesNotAbort(t *testing.T) {
	t.Parallel()

	require.False(t, templateStep{}.Abort(&Options{}))
	require.False(t, templateStep{}.Abort(&Options{Template: "cli"}))
}
