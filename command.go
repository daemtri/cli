package cli

import (
	"flag"
	"fmt"
	"github.com/duanqy/cli/internal/container"
	"github.com/duanqy/cli/internal/fsm"
	"github.com/duanqy/cli/internal/lexer"
	"github.com/duanqy/cli/internal/parser"
	"io"
	"strings"
)

type Action func(ctx Context) error

type Command interface {

}

// Cmd represents a command (or sub command) in a CLI application. It should be constructed
// by calling Command() on an app to create a top level command or by calling Command() on another
// command to create a sub command
type Cmd struct {
	// The code to execute when this command is matched
	Action Action
	// The code to execute before this command or any of its children is matched
	Before Action
	// The code to execute after this command or any of its children is matched
	After Action
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

	parent *Cmd

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
		parent:        c,
	})
}

func (c *Cmd) doInit() error {
	if c.init != nil {
		c.init(c)
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

func (c *Cmd) callBefore() error {
	if c.parent != nil {
		if err := c.parent.callBefore(); err != nil {
			return err
		}
	}
	if c.Before != nil {
		return c.Before(newContext(c))
	}
	return nil
}

func (c *Cmd) callAfter(err error) error {
	ctx := newContext(c)
	ctx.err = err
	if c.After != nil {
		if err := c.After(ctx); err != nil {
			ctx.err = err
		}
	}
	if c.parent != nil {
		return c.parent.callAfter(err)
	}
	return err
}

func (c *Cmd) run(args []string) (err error) {
	if c.helpRequested(args) {
		c.PrintLongHelp()
		c.onError(errHelpRequested)
		return nil
	}

	nargsLen := c.getOptsAndArgs(args)

	if err := c.fsm.Parse(args[:nargsLen]); err != nil {
		_, _ = fmt.Fprintf(stdErr, "error: %s\n", err.Error())
		c.PrintHelp()
		c.onError(err)
		return err
	}

	args = args[nargsLen:]
	if len(args) == 0 {
		if c.Action != nil {
			if err = c.callBefore(); err != nil {
				return err
			}
			defer func() {
				err = c.callAfter(err)
			}()
			return c.Action(newContext(c))
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
			return sub.run(args[1:])
		}
	}

	switch {
	case strings.HasPrefix(arg, "-"):
		err = fmt.Errorf("error: illegal option %s", arg)
		_, _ = fmt.Fprintln(stdErr, err.Error())
	default:
		err = fmt.Errorf("error: illegal input %s", arg)
		_, _ = fmt.Fprintln(stdErr, err.Error())
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
	_, _ = fmt.Fprintf(w, "  %s\t%s\n", s1, strings.TrimSpace(lines[0]))

	if len(lines) == 1 {
		return
	}

	for _, line := range lines[1:] {
		_, _ = fmt.Fprintf(w, "  %s\t%s\n", "", strings.TrimSpace(line))
	}
}
