package cli

import (
	"flag"
	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
	"time"
)

type Parameter interface {
	Env(key string) Parameter
	Hide() Parameter

	Bool(def bool) *bool
	BoolVar(p *bool, def bool)

	String(def string) *string
	StringVar(p *string, def string)

	Int(def int) *int
	IntVar(p *int, def int)

	Int64(def int64) *int64
	Int64Var(p *int64, def int64)

	Uint(def uint) *uint
	UintVar(p *uint, def uint)

	Uint64(def uint64) *uint64
	Uint64Var(p *uint64, def uint64)

	Float64(def float64) *float64
	Float64Var(p *float64, def float64)

	Duration(def time.Duration) *time.Duration
	DurationVar(p *time.Duration, def time.Duration)

	Var(v flag.Value)
}

type parameter struct {
	c *container.Container
}

func (pa *parameter) Env(key string) Parameter {
	pa.c.EnvVar = key
	pa.c.ValueSetFromEnv = values.SetFromEnv(pa.c.Value, pa.c.EnvVar)
	return pa
}

func (pa *parameter) Hide() Parameter {
	pa.c.HideValue = true
	return pa
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
