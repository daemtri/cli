package container

import (
	"flag"
)

// Container holds an option or an arg data
type Container struct {
	Name            string
	Desc            string
	EnvVar          string
	Names           []string
	Hidden          bool
	ValueSetFromEnv bool
	ValueSetByUser  *bool
	Value           flag.Value
	Default         interface{}
}
