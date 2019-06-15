package cli

import (
	"fmt"
	"strings"

	"github.com/jawher/mow.cli/internal/container"
)

func mkOptStrs(optName string) []string {
	res := strings.Fields(optName)
	for i, name := range res {
		prefix := "-"
		if len(name) > 1 {
			prefix = "--"
		}
		res[i] = prefix + name
	}
	return res
}

func (c *Cmd) mkOpt(opt *container.Container) {
	opt.Names = mkOptStrs(opt.Name)

	c.options = append(c.options, opt)
	for _, name := range opt.Names {
		if _, found := c.optionsIdx[name]; found {
			panic(fmt.Sprintf("duplicate option name %q", name))
		}
		c.optionsIdx[name] = opt
	}
}
