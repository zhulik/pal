package initcmd

import "github.com/urfave/cli/v3"

// Options holds the init configuration read from CLI flags/args (and later
// refined by the interactive wizard when enabled).
type Options struct {
	NoInteractive bool
	// Directory is the project path from --directory/-d. Empty means the
	// process working directory. Not a Step — CLI-only path override.
	Directory string
	// Module is the Go module path passed to `go mod init`.
	Module string
	// Template is the project template name (directory under
	// cmd/pal/templates/scaffolds). Default is "cli" when read from flags; the
	// wizard also defaults to cli.
	Template string
	// TemplateSet is true when --template was explicitly passed on the CLI.
	// Used by the wizard to skip the template step without treating the flag
	// default as a user choice.
	TemplateSet bool
	// Git controls whether init runs `git init`. Default is true when read from
	// flags (unless --no-git); the wizard also defaults to yes.
	Git bool
	// GitSet is true when --no-git was explicitly passed on the CLI.
	GitSet bool
}

// Args returns the CLI arguments equivalent to the project options
// (excluding mode flags such as --no-interactive).
func (o Options) Args() []string {
	var args []string
	if o.Module != "" {
		args = append(args, o.Module)
	}
	if o.Directory != "" {
		args = append(args, "-d", o.Directory)
	}
	if o.Template != "" && o.Template != defaultTemplate {
		args = append(args, "--"+flagTemplate, o.Template)
	}
	if !o.Git {
		args = append(args, "--"+flagNoGit)
	}
	return args
}

// OptionsFromCommand reads init options from parsed CLI flags and arguments.
func OptionsFromCommand(cmd *cli.Command) Options {
	return Options{
		NoInteractive: cmd.Bool(flagNoInteractive),
		Directory:     cmd.String(flagDirectory),
		Module:        cmd.StringArg(argModule),
		Template:      cmd.String(flagTemplate),
		TemplateSet:   cmd.IsSet(flagTemplate),
		Git:           !cmd.Bool(flagNoGit),
		GitSet:        cmd.IsSet(flagNoGit),
	}
}
