package initcmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestStepsFlagsCoveredByFlags(t *testing.T) {
	t.Parallel()

	flagNames := map[string]struct{}{}
	for _, f := range Flags() {
		for _, name := range f.Names() {
			flagNames[name] = struct{}{}
		}
	}

	for _, step := range Steps() {
		flag := step.Flag()
		require.NotNil(t, flag)
		names := flag.Names()
		require.NotEmpty(t, names, "step flag must have a name")
		_, ok := flagNames[names[0]]
		require.True(t, ok, "step flag %q must appear in Flags()", names[0])
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
		for _, name := range step.Flag().Names() {
			require.NotEqual(t, flagNoInteractive, name)
		}
	}
}

func TestForceStepFlag(t *testing.T) {
	t.Parallel()

	flag := forceStep{}.Flag()
	boolFlag, ok := flag.(*cli.BoolFlag)
	require.True(t, ok)
	require.Equal(t, "force", boolFlag.Name)
	require.Equal(t, []string{"f"}, boolFlag.Aliases)
}
