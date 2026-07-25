// Package version reports the pal CLI build version.
package version

import (
	"fmt"
	"runtime/debug"
	"time"
)

const (
	devel   = "(devel)"
	unknown = "unknown"
	shaLen  = 7
)

// String returns the CLI version for --version and the version subcommand.
//
// Prefer the Go module version when the binary was built from a tagged module
// (or a pseudo-version). For local / untagged builds (Main.Version ==
// "(devel)"), fall back to the short VCS revision and commit date, with a
// "-dirty" suffix when the working tree was dirty at build time.
func String() string {
	return FromBuildInfo(debug.ReadBuildInfo())
}

// FromBuildInfo formats a version string from Go build metadata. ok is the
// second return value of debug.ReadBuildInfo.
func FromBuildInfo(info *debug.BuildInfo, ok bool) string {
	if !ok || info == nil {
		return unknown
	}

	if v := info.Main.Version; v != "" && v != devel {
		return v
	}

	var revision, vcsTime string
	modified := false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.time":
			vcsTime = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}

	if revision == "" {
		return "devel"
	}
	if len(revision) > shaLen {
		revision = revision[:shaLen]
	}
	if modified {
		revision += "-dirty"
	}
	if vcsTime == "" {
		return revision
	}
	if t, err := time.Parse(time.RFC3339, vcsTime); err == nil {
		vcsTime = t.UTC().Format("2006-01-02")
	}
	return fmt.Sprintf("%s %s", revision, vcsTime)
}
