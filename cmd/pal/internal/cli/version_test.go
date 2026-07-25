package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zhulik/pal/cmd/pal/internal/cli"
	"github.com/zhulik/pal/cmd/pal/internal/version"
)

func TestVersion_FlagAndSubcommandMatch(t *testing.T) {
	t.Parallel()

	want := "pal version " + version.String() + "\n"

	for _, args := range [][]string{
		{"pal", "--version"},
		{"pal", "-v"},
		{"pal", "version"},
	} {
		t.Run(strings.Join(args[1:], " "), func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			cmd := cli.New()
			cmd.Writer = &out

			err := cmd.Run(context.Background(), args)
			require.NoError(t, err)
			require.Equal(t, want, out.String())
		})
	}
}
