package cli

import (
	"time"
)

type context struct {
	cmd    *Cmd
	parent *context
	err    error
}

func (c context) Prompt(name string, desc string) Value {
	panic("implement me")
}

func (c context) Error() error {
	return c.err
}

func (c context) Argument(name string) Value {
	panic("implement me")
}

func (c context) Option(name string) Value {
	panic("implement me")
}

func newContext(cmd *Cmd) *context {
	return &context{cmd: cmd}
}

var _ Context = &context{}

type Value interface {
	Bool() bool
	Duration() time.Duration
	// 输入不可见
	Password() string

	String() string
	Int() int
	Int64() int64
	Uint() uint
	Uint64() uint64
	Float64() float64

	StringSlice() []string
	IntSlice() []int
	Int64Slice() []int64
	UintSlice() []uint
	Uint64Slice() []uint64
	Float64Slice() []float64
}


type Context interface {
	Argument(name string) Value
	Option(name string) Value
	Prompt(name string, desc string) Value
	Error() error
}
