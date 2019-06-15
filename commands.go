package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/flow"
	"github.com/jawher/mow.cli/internal/fsm"
	"github.com/jawher/mow.cli/internal/lexer"
	"github.com/jawher/mow.cli/internal/parser"
	"github.com/jawher/mow.cli/internal/values"
)

// Cmd represents a command (or sub command) in a CLI application. It should be constructed
// by calling Command() on an app to create a top level command or by calling Command() on another
// command to create a sub command
type Cmd struct {
	// The code to execute when this command is matched
	Action func()
	// The code to execute before this command or any of its children is matched
	Before func()
	// The code to execute after this command or any of its children is matched
	After func()
	// The command options and arguments
	Spec string
	// The command long description to be shown when help is requested
	LongDesc string
	// The command error handling strategy
	ErrorHandling flag.ErrorHandling

	init    CmdInitializer
	name    string
	aliases []string
	desc    string

	commands   []*Cmd
	options    []*container.Container
	optionsIdx map[string]*container.Container
	args       []*container.Container
	argsIdx    map[string]*container.Container

	parents []string

	fsm *fsm.State
}

// CmdInitializer is a function that configures a command by adding options, arguments, a spec, sub commands and the code
// to execute when the command is called
type CmdInitializer func(*Cmd)

// Command adds a new (sub) command to c where name is the command name (what you type in the console),
// description is what would be shown in the help messages, e.g.:
//
//	Usage: git [OPTIONS] COMMAND [arg...]
//
//	Commands:
//	  $name	$desc
//
// the last argument, init, is a function that will be called by mow.cli to further configure the created
// (sub) command, e.g. to add options, arguments and the code to execute
func (c *Cmd) Command(name, desc string, init CmdInitializer) {
	aliases := strings.Fields(name)
	c.commands = append(c.commands, &Cmd{
		ErrorHandling: c.ErrorHandling,
		name:          aliases[0],
		aliases:       aliases,
		desc:          desc,
		init:          init,
		commands:      []*Cmd{},
		options:       []*container.Container{},
		optionsIdx:    map[string]*container.Container{},
		args:          []*container.Container{},
		argsIdx:       map[string]*container.Container{},
	})
}

func (c *Cmd) doInit() error {
	if c.init != nil {
		c.init(c)
	}

	parents := append(c.parents, c.name)

	for _, sub := range c.commands {
		sub.parents = parents
	}

	if len(c.Spec) == 0 {
		if len(c.options) > 0 {
			c.Spec = "[OPTIONS] "
		}
		for _, arg := range c.args {
			c.Spec += arg.Name + " "
		}
	}

	tokens, err := lexer.Tokenize(c.Spec)
	if err != nil {
		return err
	}

	params := parser.Params{
		Spec:       c.Spec,
		Options:    c.options,
		OptionsIdx: c.optionsIdx,
		Args:       c.args,
		ArgsIdx:    c.argsIdx,
	}
	s, err := parser.Parse(tokens, params)
	if err != nil {
		return err
	}
	c.fsm = s
	return nil
}

func (c *Cmd) onError(err error) {
	if err == errHelpRequested || err == errVersionRequested {
		if c.ErrorHandling == flag.ExitOnError {
			exiter(0)
		}
		return
	}

	switch c.ErrorHandling {
	case flag.ExitOnError:
		exiter(2)
	case flag.PanicOnError:
		panic(err)
	}

}

// PrintHelp prints the command's help message.
// In most cases the library users won't need to call this method, unless
// a more complex validation is needed
func (c *Cmd) PrintHelp() {
	c.printHelp(false)
}

// PrintLongHelp prints the command's help message using the command long description if specified.
// In most cases the library users won't need to call this method, unless
// a more complex validation is needed
func (c *Cmd) PrintLongHelp() {
	c.printHelp(true)
}

func (c *Cmd) printHelp(longDesc bool) {
	full := append(c.parents, c.name)
	path := strings.Join(full, " ")
	_,_ = fmt.Fprintf(stdErr, "\nUsage: %s", path)

	spec := strings.TrimSpace(c.Spec)
	if len(spec) > 0 {
		_,_ = fmt.Fprintf(stdErr, " %s", spec)
	}

	if len(c.commands) > 0 {
		_,_ =fmt.Fprint(stdErr, " COMMAND [arg...]")
	}
	_,_ =fmt.Fprint(stdErr, "\n\n")

	desc := c.desc
	if longDesc && len(c.LongDesc) > 0 {
		desc = c.LongDesc
	}
	if len(desc) > 0 {
		_,_ =fmt.Fprintf(stdErr, "%s\n", desc)
	}

	w := tabwriter.NewWriter(stdErr, 15, 1, 3, ' ', 0)

	if len(c.args) > 0 {
		_,_ =fmt.Fprint(w, "\t\nArguments:\t\n")
		for _, arg := range c.args {
			if arg.HideValue {
				continue
			}
			var (
				env   = formatEnvVarsForHelp(arg.EnvVar)
				value = formatValueForHelp(arg.Value)
			)
			printTabbedRow(w, arg.Name, joinStrings(arg.Desc, env, value))
		}
	}

	if len(c.options) > 0 {
		_,_ =fmt.Fprint(w, "\t\nOptions:\t\n")
		for _, opt := range c.options {
			if opt.HideValue {
				continue
			}
			var (
				optNames = formatOptNamesForHelp(opt)
				env      = formatEnvVarsForHelp(opt.EnvVar)
				value    = formatValueForHelp(opt.Value)
			)
			printTabbedRow(w, optNames, joinStrings(opt.Desc, env, value))
		}
	}

	if len(c.commands) > 0 {
		_,_ =fmt.Fprint(w, "\t\nCommands:\t\n")

		for _, c := range c.commands {
			_,_ =fmt.Fprintf(w, "  %s\t%s\n", strings.Join(c.aliases, ", "), c.desc)
		}
	}

	if len(c.commands) > 0 {
		_,_ =fmt.Fprintf(w, "\t\nRun '%s COMMAND --help' for more information on a command.\n", path)
	}

	_ = w.Flush()
}

func formatOptNamesForHelp(o *container.Container) string {
	short, long := "", ""

	for _, n := range o.Names {
		if len(n) == 2 && short == "" {
			short = n
		}

		if len(n) > 2 && long == "" {
			long = n
		}
	}

	switch {
	case short != "" && long != "":
		return fmt.Sprintf("%s, %s", short, long)
	case short != "":
		return short
	case long != "":
		// 2 spaces instead of the short option (-x), one space for the comma (,) and one space for the after comma blank
		return fmt.Sprintf("    %s", long)
	default:
		return ""
	}
}

func formatValueForHelp(v flag.Value) string {
	if dv, ok := v.(values.DefaultValued); ok {
		if dv.IsDefault() {
			return ""
		}
	}

	return fmt.Sprintf("(default %s)", v.String())
}

func formatEnvVarsForHelp(envVars string) string {
	if strings.TrimSpace(envVars) == "" {
		return ""
	}
	vars := strings.Fields(envVars)
	res := "(env"
	sep := " "
	for i, v := range vars {
		if i > 0 {
			sep = ", "
		}
		res += fmt.Sprintf("%s$%s", sep, v)
	}
	res += ")"
	return res
}

func (c *Cmd) parse(args []string, entry, inFlow, outFlow *flow.Step) error {
	if c.helpRequested(args) {
		c.PrintLongHelp()
		c.onError(errHelpRequested)
		return nil
	}

	nargsLen := c.getOptsAndArgs(args)

	if err := c.fsm.Parse(args[:nargsLen]); err != nil {
		fmt.Fprintf(stdErr, "Error: %s\n", err.Error())
		c.PrintHelp()
		c.onError(err)
		return err
	}

	newInFlow := &flow.Step{
		Do:     c.Before,
		Error:  outFlow,
		Desc:   fmt.Sprintf("%s.Before", c.name),
		Exiter: exiter,
	}
	inFlow.Success = newInFlow

	newOutFlow := &flow.Step{
		Do:      c.After,
		Success: outFlow,
		Error:   outFlow,
		Desc:    fmt.Sprintf("%s.After", c.name),
		Exiter:  exiter,
	}

	args = args[nargsLen:]
	if len(args) == 0 {
		if c.Action != nil {
			newInFlow.Success = &flow.Step{
				Do:      c.Action,
				Success: newOutFlow,
				Error:   newOutFlow,
				Desc:    fmt.Sprintf("%s.Action", c.name),
				Exiter:  exiter,
			}

			entry.Run(nil)
			return nil
		}
		c.PrintHelp()
		c.onError(nil)
		return nil
	}

	arg := args[0]
	for _, sub := range c.commands {
		if sub.isAlias(arg) {
			if err := sub.doInit(); err != nil {
				panic(err)
			}
			return sub.parse(args[1:], entry, newInFlow, newOutFlow)
		}
	}

	var err error
	switch {
	case strings.HasPrefix(arg, "-"):
		err = fmt.Errorf("Error: illegal option %s", arg)
		fmt.Fprintln(stdErr, err.Error())
	default:
		err = fmt.Errorf("Error: illegal input %s", arg)
		fmt.Fprintln(stdErr, err.Error())
	}
	c.PrintHelp()
	c.onError(err)
	return err

}

func (c *Cmd) helpRequested(args []string) bool {
	return c.isFlagSet(args, []string{"-h", "--help"})
}

func (c *Cmd) isFlagSet(args []string, searchArgs []string) bool {
	if len(args) == 0 {
		return false
	}

	arg := args[0]
	for _, searchArg := range searchArgs {
		if arg == searchArg {
			return true
		}
	}
	return false
}

func (c *Cmd) getOptsAndArgs(args []string) int {
	consumed := 0

	for _, arg := range args {
		for _, sub := range c.commands {
			if sub.isAlias(arg) {
				return consumed
			}
		}
		consumed++
	}
	return consumed
}

func (c *Cmd) isAlias(arg string) bool {
	for _, alias := range c.aliases {
		if arg == alias {
			return true
		}
	}
	return false
}

func joinStrings(parts ...string) string {
	res := ""
	for _, part := range parts {
		s := strings.TrimSpace(part)
		if s == "" {
			continue
		}
		if res != "" {
			res += " "
		}
		res += part
	}
	return res
}

func printTabbedRow(w io.Writer, s1 string, s2 string) {
	lines := strings.Split(s2, "\n")
	_,_ = fmt.Fprintf(w, "  %s\t%s\n", s1, strings.TrimSpace(lines[0]))

	if len(lines) == 1 {
		return
	}

	for _, line := range lines[1:] {
		_,_ = fmt.Fprintf(w, "  %s\t%s\n", "", strings.TrimSpace(line))
	}
}
