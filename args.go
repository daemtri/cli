package cli

import (
	"flag"
	"fmt"

	"github.com/jawher/mow.cli/internal/lexer"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
)

// BoolArg describes a boolean argument
type BoolArg struct {
	Parameter
	// The argument's initial value
	Value bool
}

func (a BoolArg) value(into *bool) (flag.Value, *bool) {
	if into == nil {
		into = new(bool)
	}
	return values.NewBool(into, a.Value), into
}

// StringArg describes a string argument
type StringArg struct {
	Parameter
	// The argument's initial value
	Value string
}

func (a StringArg) value(into *string) (flag.Value, *string) {
	if into == nil {
		into = new(string)
	}
	return values.NewString(into, a.Value), into
}

// IntArg describes an int argument
type IntArg struct {
	Parameter
	// The argument's initial value
	Value int
}

func (a IntArg) value(into *int) (flag.Value, *int) {
	if into == nil {
		into = new(int)
	}
	return values.NewInt(into, a.Value), into
}

// Float64Arg describes an float64 argument
type Float64Arg struct {
	Parameter
	// The argument's initial value
	Value float64
}

func (a Float64Arg) value(into *float64) (flag.Value, *float64) {
	if into == nil {
		into = new(float64)
	}
	return values.NewFloat64(into, a.Value), into
}

// StringsArg describes a string slice argument
type StringsArg struct {
	Parameter
	// The argument's initial value
	Value []string
}

func (a StringsArg) value(into *[]string) (flag.Value, *[]string) {
	if into == nil {
		into = new([]string)
	}
	return values.NewStrings(into, a.Value), into
}

// IntsArg describes an int slice argument
type IntsArg struct {
	Parameter
	// The argument's initial value
	Value []int
}

func (a IntsArg) value(into *[]int) (flag.Value, *[]int) {
	if into == nil {
		into = new([]int)
	}
	return values.NewInts(into, a.Value), into
}

// Floats64Arg describes an int slice argument
type Floats64Arg struct {
	Parameter
	// The argument's initial value
	Value []float64
}

func (a Floats64Arg) value(into *[]float64) (flag.Value, *[]float64) {
	if into == nil {
		into = new([]float64)
	}
	return values.NewFloats64(into, a.Value), into
}

// VarArg describes an argument where the type and format of the value is controlled by the developer
type VarArg struct {
	Parameter
	// A value implementing the flag.Value type (will hold the final value)
	Value flag.Value
}

func (a VarArg) value() flag.Value {
	return a.Value
}


func (c *Cmd) mkArg(arg *container.Container) {
	if !validArgName(arg.Name) {
		panic(fmt.Sprintf("invalid argument name %q: must be in all caps", arg.Name))
	}
	if _, found := c.argsIdx[arg.Name]; found {
		panic(fmt.Sprintf("duplicate argument name %q", arg.Name))
	}

	c.args = append(c.args, arg)
	c.argsIdx[arg.Name] = arg
}

func validArgName(n string) bool {
	tokens, err := lexer.Tokenize(n)
	if err != nil {
		return false
	}
	if len(tokens) != 1 {
		return false
	}

	return tokens[0].Typ == lexer.TTArg
}
