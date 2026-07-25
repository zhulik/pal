package initcmd

import "github.com/urfave/cli/v3"

// Options holds the init configuration read from CLI flags (and later refined
// by the interactive wizard when enabled).
type Options struct {
	NoInteractive bool
	Force         bool
	// Git controls whether init runs `git init`. Default is true when read from
	// flags (unless --no-git); the wizard also defaults to yes.
	Git bool
}

// Args returns the CLI flag arguments equivalent to the project options
// (excluding mode flags such as --no-interactive).
func (o Options) Args() []string {
	var args []string
	if o.Force {
		args = append(args, "--force")
	}
	if !o.Git {
		args = append(args, "--"+flagNoGit)
	}
	return args
}

// OptionsFromCommand reads init options from parsed CLI flags.
func OptionsFromCommand(cmd *cli.Command) Options {
	return Options{
		NoInteractive: cmd.Bool(flagNoInteractive),
		Force:         cmd.Bool("force"),
		Git:           !cmd.Bool(flagNoGit),
	}
}
