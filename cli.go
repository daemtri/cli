package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/duanqy/cli/internal/container"
)

// App represents the structure of a CLI app. It should be constructed using the App() function
type App struct {
	*Cmd
	version *cliVersion
}

type cliVersion struct {
	version string
	option  *container.Container
}

// App creates a new and empty CLI app configured with the passed name and description.
//
// name and description will be used to construct the help message for the app:
//
//	Usage: $name [OPTIONS] COMMAND [arg...]
//
//	$desc
func NewApp(name, desc string) *App {
	return &App{
		Cmd: &Cmd{
			name:          name,
			desc:          desc,
			optionsIdx:    map[string]*container.Container{},
			argsIdx:       map[string]*container.Container{},
			ErrorHandling: flag.ExitOnError,
		},
	}
}

// Version sets the version string of the CLI app together with the options that can be used to trigger
// printing the version string via the CLI.
//
//	Usage: appName --$name
//	$version
func (a *App) Version(name, version string) {
	a.Option(name, "Show the version and exit").Bool(false)
	names := mkOptStrs(name)
	option := a.optionsIdx[names[0]]
	a.version = &cliVersion{version, option}
}

func (a *App) run(args []string) error {
	// We overload Cmd.parse() and handle cases that only apply to the CLI command, like versioning
	// After that, we just call Cmd.parse() for the default behavior
	if a.versionSetAndRequested(args) {
		a.PrintVersion()
		a.onError(errVersionRequested)
		return nil
	}
	return a.Cmd.run(args)
}

func (a *App) versionSetAndRequested(args []string) bool {
	return a.version != nil && a.isFlagSet(args, a.version.option.Names)
}

/*
PrintVersion prints the CLI app's version.
In most cases the library users won't need to call this method, unless
a more complex validation is needed.
*/
func (a *App) PrintVersion() {
	_, _ = fmt.Fprintln(stdErr, a.version.version)
}

/*
Run uses the app configuration (specs, commands, ...) to parse the args slice
and to execute the matching command.

In case of an incorrect usage, and depending on the configured ErrorHandling policy,
it may return an error, panic or exit
*/
func (a *App) Run(args []string) error {
	if err := a.doInit(); err != nil {
		panic(err)
	}
	return a.run(args[1:])
}

// ActionCommand is a convenience function to configure a command with an action.
//
// cmd.ActionCommand(_, _, myFunc) is equivalent to cmd.Command(_, _, func(cmd *cli.Cmd) { cmd.Action = myFunc })
func ActionCommand(action func(ctx Context) error) CmdInitializer {
	return func(cmd *Cmd) {
		cmd.Action = action
	}
}

var exiter = func(code int) {
	os.Exit(code)
}

var (
	stdOut io.Writer = os.Stdout
	stdErr io.Writer = os.Stderr
)
