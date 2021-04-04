// Package shell handles interactions with the OS shell
package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
)

// Type shell represents a shell command
const Type = "shell"

// Command represents the function that builds the exec.Cmd to be used in shell commands.
type command = func(name string, arg ...string) *exec.Cmd

// Execute represents a way to execute shell commands with values piped to them.
type Execute struct {
	Command command
}

// ExecuteWithValue executes a shell command with a value piped to it.
// If it exits with code 0, no error is returned and the stdout is captured and returned.
// If it exits with code 1, an error is returned and the stderr is captured and returned.
// If the timeout is reached, an error is returned.
func (e *Execute) ExecuteWithValue(definition *hook.Definition, value string) (string, error) {
	if definition.Shell == nil {
		return "", fmt.Errorf("Missing required 'shell' configuration on hook definition")
	}
	// Build command string with value piped into it
	cmd := e.Command(definition.Shell.Entrypoint, definition.Shell.Command...)

	// Set up byte buffer to write values to stdin
	inb := bytes.Buffer{}
	// No need to catch error, doesn't produce error, instead it panics if buffer too large
	inb.WriteString(value)
	cmd.Stdin = &inb

	// Set up byte buffers to read stdout and stderr
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	// Start command
	err := cmd.Start()
	if err != nil {
		return "", err
	}

	// Set up channel to wait for command to finish
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Set up a timeout, after which if the command hasn't finished it will be stopped
	timeoutListener := time.After(time.Duration(definition.Timeout) * time.Millisecond)

	select {
	case <-timeoutListener:
		cmd.Process.Kill()
		return "", fmt.Errorf("Entrypoint '%s', command '%s' timed out", definition.Shell.Entrypoint, definition.Shell.Command)
	case err = <-done:
		if err != nil {
			return "", fmt.Errorf("%v: %s", err, errb.String())
		}
	}
	return outb.String(), nil
}

// GetType returns the shell executer type
func (e *Execute) GetType() string {
	return Type
}
