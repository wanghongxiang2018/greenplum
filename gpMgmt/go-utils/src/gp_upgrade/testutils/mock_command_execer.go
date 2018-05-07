package testutils

import (
	"strings"
	"sync"

	"gp_upgrade/helpers"
)

type FakeCommandExecer struct {
	command        string
	args           []string
	calls          []string
	numInvocations int

	mu      sync.Mutex
	output  helpers.Command
	trigger chan struct{}
}

func (c *FakeCommandExecer) SetOutput(command helpers.Command) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.output = command
}

func (c *FakeCommandExecer) SetTrigger(trigger chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.trigger = trigger
}

func (c *FakeCommandExecer) Calls() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.calls
}

func (c *FakeCommandExecer) Command() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.command
}

func (c *FakeCommandExecer) Args() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.args
}

func (c *FakeCommandExecer) GetNumInvocations() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.numInvocations
}

func (c *FakeCommandExecer) Exec(command string, args ...string) helpers.Command {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.numInvocations++

	c.command = command
	c.args = args

	calledWith := append([]string{command}, args...)
	c.calls = append(c.calls, strings.Join(calledWith, " "))

	if c.trigger != nil {
		<-c.trigger
	}

	return c.output
}

type FakeCommand struct {
	Err     chan error
	Out     chan []byte
	Trigger chan struct{}
}

func (c *FakeCommand) Output() ([]byte, error) {
	var err error
	var out []byte

	if len(c.Err) != 0 {
		err = <-c.Err
	}

	if len(c.Out) != 0 {
		out = <-c.Out
	}

	return out, err
}

func (c *FakeCommand) CombinedOutput() ([]byte, error) {
	var err error
	var out []byte

	if len(c.Err) != 0 {
		err = <-c.Err
	}

	if len(c.Out) != 0 {
		out = <-c.Out
	}

	return out, err
}

func (c *FakeCommand) Start() error {
	var err error

	if len(c.Err) != 0 {
		err = <-c.Err
	}

	return err
}

func (c *FakeCommand) Run() error {
	var err error

	if len(c.Err) != 0 {
		err = <-c.Err
	}

	if c.Trigger != nil {
		<-c.Trigger
	}

	return err
}
