/*
Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.

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
	"os"
	"os/exec"
	"path"
	"time"
)

const (
	entrypoint = "python"
)

type command = func(name string, arg ...string) *exec.Cmd

func NewAlgorithmPython() *Python {
	return &Python{
		Command: exec.Command,
		Getwd:   os.Getwd,
	}
}

// Python is an implementation of an algorithm runner that runs algorithms using Python commands
type Python struct {
	Command command
	Getwd   func() (dir string, err error)
}

// RunAlgorithmWithValue runs an algorithm at the path provided, passing through the value provided
func (r *Python) RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error) {

	wd, err := r.Getwd()
	if err != nil {
		return "", err
	}

	cmd := r.Command(entrypoint, path.Join(wd, algorithmPath))

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
	err = cmd.Start()
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
		return "", fmt.Errorf("entrypoint '%s', command '%s' timed out", entrypoint, algorithmPath)
	case err = <-done:
		if err != nil {
			return "", fmt.Errorf("%v: %s", err, errb.String())
		}
	}
	return outb.String(), nil
}
