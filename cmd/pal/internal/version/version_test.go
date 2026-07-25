package version_test

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zhulik/pal/cmd/pal/internal/version"
)

func TestFromBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "missing build info",
			ok:   false,
			want: "unknown",
		},
		{
			name: "nil build info",
			ok:   true,
			want: "unknown",
		},
		{
			name: "tagged module version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			ok:   true,
			want: "v1.2.3",
		},
		{
			name: "pseudo-version",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v0.0.0-20260725143348-abcdef012345"},
			},
			ok:   true,
			want: "v0.0.0-20260725143348-abcdef012345",
		},
		{
			name: "devel with revision and time",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef0123456789"},
					{Key: "vcs.time", Value: "2026-07-25T12:34:56Z"},
				},
			},
			ok:   true,
			want: "abcdef0 2026-07-25",
		},
		{
			name: "devel dirty",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef0123456789"},
					{Key: "vcs.time", Value: "2026-07-25T12:34:56Z"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			ok:   true,
			want: "abcdef0-dirty 2026-07-25",
		},
		{
			name: "devel revision only",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef0123456789"},
				},
			},
			ok:   true,
			want: "abcdef0",
		},
		{
			name: "devel without vcs",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: "devel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, version.FromBuildInfo(tt.info, tt.ok))
		})
	}
}

func TestString_nonEmpty(t *testing.T) {
	t.Parallel()
	require.NotEmpty(t, version.String())
	require.NotEqual(t, "unknown", version.String())
}
