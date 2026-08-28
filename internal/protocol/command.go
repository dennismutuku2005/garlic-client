package protocol

import (
	"fmt"
	"strings"
)

// Command represents an API command to be sent to RouterOS.
type Command struct {
	Name string   // The command path, e.g., "/ip/address/print"
	Args []string // Parameters and filters, e.g., "=interface=ether1", "?disabled=false"
	Tag  string   // Optional correlation tag
}

// NewCommand creates a new command with the specified command path.
func NewCommand(name string) *Command {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	return &Command{
		Name: name,
	}
}

// Arg adds a general key-value argument to the command.
// It will be formatted as "=key=value".
func (c *Command) Arg(key, value string) *Command {
	c.Args = append(c.Args, fmt.Sprintf("=%s=%s", key, value))
	return c
}

// Attr adds an argument directly, useful for raw command parameters.
func (c *Command) Attr(val string) *Command {
	c.Args = append(c.Args, val)
	return c
}

// Filter adds a query filter to the command.
// It will be formatted as "?key=value".
func (c *Command) Filter(key, value string) *Command {
	c.Args = append(c.Args, fmt.Sprintf("?%s=%s", key, value))
	return c
}

// Proplist specifies the properties/fields to return from the command.
// It will be formatted as "=.proplist=field1,field2,...".
func (c *Command) Proplist(fields ...string) *Command {
	if len(fields) > 0 {
		c.Args = append(c.Args, fmt.Sprintf("=.proplist=%s", strings.Join(fields, ",")))
	}
	return c
}

// WithTag sets the correlation tag for async queries.
func (c *Command) WithTag(tag string) *Command {
	c.Tag = tag
	return c
}

// Sentence returns the string slice representation of the command sentence.
func (c *Command) Sentence() []string {
	sentence := make([]string, 0, 1+len(c.Args)+1)
	sentence = append(sentence, c.Name)
	sentence = append(sentence, c.Args...)
	if c.Tag != "" {
		sentence = append(sentence, ".tag="+c.Tag)
	}
	return sentence
}
