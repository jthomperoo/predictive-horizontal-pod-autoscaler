/*
Copyright 2021 The Predictive Horizontal Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package algorithm

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

const (
	entrypoint = "python"
)

// Runner defines an algorithm runner, allowing algorithms to be run
type Runner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int64) (string, error)
}

type command = func(name string, arg ...string) *exec.Cmd

// Run is an implementation of an algorithm runner that uses CPA executers to run shell commands
type Run struct {
	Command command
}

// RunAlgorithmWithValue runs an algorithm at the path provided, passing through the value provided
func (r *Run) RunAlgorithmWithValue(algorithmPath string, value string, timeout int64) (string, error) {
	// Build command string with value piped into it
	cmd := r.Command(entrypoint, algorithmPath)

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
	timeoutListener := time.After(time.Duration(timeout) * time.Millisecond)

	select {
	case <-timeoutListener:
		cmd.Process.Kill()
		return "", fmt.Errorf("Algorithm at path '%s' with entrypoint '%s' timed out", algorithmPath, entrypoint)
	case err = <-done:
		if err != nil {
			return "", err
		}
	}
	return outb.String(), nil
}
