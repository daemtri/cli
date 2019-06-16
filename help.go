package cli

import (
	"flag"
	"fmt"
	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
	"strings"
	"text/tabwriter"
)

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

func (c *Cmd) fullPath() string  {
	if c.parent != nil {
		return c.parent.fullPath() + " " + c.name
	}
	return c.name
}

func (c *Cmd) printHelp(longDesc bool) {
	path := c.fullPath()
	_, _ = fmt.Fprintf(stdErr, "\nUsage: %s", path)

	spec := strings.TrimSpace(c.Spec)
	if len(spec) > 0 {
		_, _ = fmt.Fprintf(stdErr, " %s", spec)
	}

	if len(c.commands) > 0 {
		_, _ = fmt.Fprint(stdErr, " COMMAND [arg...]")
	}
	_, _ = fmt.Fprint(stdErr, "\n\n")

	desc := c.desc
	if longDesc && len(c.LongDesc) > 0 {
		desc = c.LongDesc
	}
	if len(desc) > 0 {
		_, _ = fmt.Fprintf(stdErr, "%s\n", desc)
	}

	w := tabwriter.NewWriter(stdErr, 15, 1, 3, ' ', 0)

	if len(c.args) > 0 {
		_, _ = fmt.Fprint(w, "\t\nArguments:\t\n")
		for _, arg := range c.args {
			if arg.Hidden {
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
		_, _ = fmt.Fprint(w, "\t\nOptions:\t\n")
		for _, opt := range c.options {
			if opt.Hidden {
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
		_, _ = fmt.Fprint(w, "\t\nCommands:\t\n")

		for _, c := range c.commands {
			_, _ = fmt.Fprintf(w, "  %s\t%s\n", strings.Join(c.aliases, ", "), c.desc)
		}
	}

	if len(c.commands) > 0 {
		_, _ = fmt.Fprintf(w, "\t\nRun '%s COMMAND --help' for more information on a command.\n", path)
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