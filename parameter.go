package cli

import (
	"flag"
	"fmt"
	"github.com/duanqy/cli/internal/container"
	"github.com/duanqy/cli/internal/lexer"
	"github.com/duanqy/cli/internal/values"
	"strings"
	"time"
)

type Parameter interface {
	Env(key string, deprecated ...string) Parameter
	Hide() Parameter
	Editor(path string) Parameter
	Deprecated(phrases string) Parameter
	Validate(func(string) bool) Parameter

	Password() *string
	PasswordVar(p *string)

	Bool(def bool) *bool
	BoolVar(p *bool, def bool)

	String(def string) *string
	StringVar(p *string, def string)

	StringSlice(def []string) *[]string
	StringSliceVar(p *[]string, def []string)

	Int(def int) *int
	IntVar(p *int, def int)

	IntSlice(def []int) *[]int
	IntSliceVar(p *[]int, def []int)

	Int64(def int64) *int64
	Int64Var(p *int64, def int64)

	Uint(def uint) *uint
	UintVar(p *uint, def uint)

	UintSlice(def []uint) *[]uint
	UintSliceVar(p *[]uint, def []uint)

	Uint64(def uint64) *uint64
	Uint64Var(p *uint64, def uint64)

	Uint64Slice(def []uint64) *[]uint64
	Uint64SliceVar(p *[]uint64, def []uint64)

	Float64(def float64) *float64
	Float64Var(p *float64, def float64)

	Float64Slice(def []float64) *[]float64
	Float64SliceVar(p *[]float64, def []float64)

	Duration(def time.Duration) *time.Duration
	DurationVar(p *time.Duration, def time.Duration)

	Var(v flag.Value)
}

type parameter struct {
	c *container.Container
}

func (pa *parameter) Env(key string, deprecated ...string) Parameter {
	pa.c.EnvVar = key
	pa.c.ValueSetFromEnv = values.SetFromEnv(pa.c.Value, pa.c.EnvVar)
	return pa
}

func (pa *parameter) Deprecated(phrases string) Parameter {
	panic("implement me")
}

func (pa *parameter) Editor(path string) Parameter {
	panic("implement me")
}

func (pa *parameter) Password() *string {
	panic("implement me")
}

func (pa *parameter) PasswordVar(p *string) {
	panic("implement me")
}

func (pa *parameter) Validate(func(string) bool) Parameter {
	panic("implement me")
}


func (pa *parameter) Hide() Parameter {
	pa.c.Hidden = true
	return pa
}

func (pa *parameter) StringSlice(def []string) *[]string {
	panic("implement me")
}

func (pa *parameter) StringSliceVar(p *[]string, def []string) {
	panic("implement me")
}

func (pa *parameter) IntSlice(def []int) *[]int {
	panic("implement me")
}

func (pa *parameter) IntSliceVar(p *[]int, def []int) {
	panic("implement me")
}

func (pa *parameter) UintSlice(def []uint) *[]uint {
	panic("implement me")
}

func (pa *parameter) UintSliceVar(p *[]uint, def []uint) {
	panic("implement me")
}

func (pa *parameter) Uint64Slice(def []uint64) *[]uint64 {
	panic("implement me")
}

func (pa *parameter) Uint64SliceVar(p *[]uint64, def []uint64) {
	panic("implement me")
}

func (pa *parameter) Float64Slice(def []float64) *[]float64 {
	panic("implement me")
}

func (pa *parameter) Float64SliceVar(p *[]float64, def []float64) {
	panic("implement me")
}

func (pa *parameter) Var(v flag.Value) {
	pa.c.Value = v
}

func (pa *parameter) Bool(def bool) *bool {
	into := new(bool)
	pa.BoolVar(into, def)
	return into
}

func (pa *parameter) BoolVar(p *bool, def bool) {
	pa.c.Value = values.NewBool(p, def)
}

func (pa *parameter) String(def string) *string {
	into := new(string)
	pa.StringVar(into, def)
	return into
}

func (pa *parameter) StringVar(p *string, def string) {
	pa.c.Value = values.NewString(p, def)
}

func (pa *parameter) Int(def int) *int {
	into := new(int)
	pa.IntVar(into, def)
	return into
}

func (pa *parameter) IntVar(p *int, def int) {
	pa.c.Value = values.NewInt(p, def)
}

func (pa *parameter) Int64(def int64) *int64 {
	into := new(int64)
	pa.Int64Var(into, def)
	return into
}

func (pa *parameter) Int64Var(p *int64, def int64) {
	pa.c.Value = values.NewInt64(p, def)
}

func (pa parameter) Uint(def uint) *uint {
	into := new(uint)
	pa.UintVar(into, def)
	return into
}

func (pa *parameter) UintVar(p *uint, def uint) {
	pa.c.Value = values.NewUint(p, def)
}

func (pa *parameter) Uint64(def uint64) *uint64 {
	into := new(uint64)
	pa.Uint64Var(into, def)
	return into
}

func (pa *parameter) Uint64Var(p *uint64, def uint64) {
	pa.c.Value = values.NewUint64(p, def)
}

func (pa *parameter) Float64(def float64) *float64 {
	into := new(float64)
	pa.Float64Var(into, def)
	return into
}

func (pa *parameter) Float64Var(p *float64, def float64) {
	pa.c.Value = values.NewFloat64(p, def)
}

func (pa *parameter) Duration(def time.Duration) *time.Duration {
	into := new(time.Duration)
	pa.DurationVar(into, def)
	return into
}

func (pa *parameter) DurationVar(p *time.Duration, def time.Duration) {
	pa.c.Value = values.NewDuration(p, def)
}

func (c *Cmd) Option(name, desc string) Parameter {
	into := new(string)
	param := &container.Container{Name: name, Desc: desc, Value: values.NewString(into, "")}
	c.mkOpt(param)
	return &parameter{c: param}
}

func (c *Cmd) Argument(name, desc string) Parameter {
	into := new(string)
	param := &container.Container{Name: name, Desc: desc, Value: values.NewString(into, "")}
	c.mkArg(param)
	return &parameter{c: param}
}

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
