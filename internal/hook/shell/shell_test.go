//go:build unit
// +build unit

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

package shell_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook/shell"
)

type command func(name string, arg ...string) *exec.Cmd

type process func(t *testing.T)

func TestShellProcess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}

	processName := strings.Split(os.Args[3], "=")[1]
	process := processes[processName]

	if process == nil {
		t.Errorf("Process %s not found", processName)
		os.Exit(1)
	}

	process(t)

	// Process should call os.Exit itself, if not exit with error
	os.Exit(1)
}

func fakeExecCommandAndStart(name string, process process) command {
	processes[name] = process
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
		cmd.Start()
		return cmd
	}
}

func fakeExecCommand(name string, process process) command {
	processes[name] = process
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1"}
		return cmd
	}
}

type test struct {
	description string
	expectedErr error
	expected    string
	definition  *hook.Definition
	pipeValue   string
	command     command
}

var tests []test

var processes map[string]process

func TestMain(m *testing.M) {
	processes = map[string]process{}
	tests = []test{
		{
			"Missing shell method configuration",
			errors.New(`Missing required 'shell' configuration on hook definition`),
			"",
			&hook.Definition{
				Type: "shell",
			},
			"test",
			exec.Command,
		},
		{
			"Successful shell command",
			nil,
			"test std out",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 100,
				Shell: &hook.Shell{
					Command:    []string{"command"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommand("success", func(t *testing.T) {
				stdinb, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
					os.Exit(1)
				}

				stdin := string(stdinb)
				entrypoint := strings.TrimSpace(os.Args[4])
				command := strings.TrimSpace(os.Args[5])

				// Check entrypoint is correct
				if !cmp.Equal(entrypoint, "/bin/sh") {
					fmt.Fprintf(os.Stderr, "entrypoint mismatch (-want +got):\n%s", cmp.Diff("/bin/sh", entrypoint))
					os.Exit(1)
				}

				// Check command is correct
				if !cmp.Equal(command, "command") {
					fmt.Fprintf(os.Stderr, "command mismatch (-want +got):\n%s", cmp.Diff("command", command))
					os.Exit(1)
				}

				// Check piped value in is correct
				if !cmp.Equal(stdin, "pipe value") {
					fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff("pipe value", stdin))
					os.Exit(1)
				}

				fmt.Fprint(os.Stdout, "test std out")
				os.Exit(0)
			}),
		},
		{
			"Successful shell command, multiple args",
			nil,
			"test std out",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 100,
				Shell: &hook.Shell{
					Command:    []string{"command", "arg1"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommand("multiple-success", func(t *testing.T) {
				stdinb, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, err.Error())
					os.Exit(1)
				}

				stdin := string(stdinb)
				entrypoint := strings.TrimSpace(os.Args[4])
				command := strings.TrimSpace(strings.Join(os.Args[5:len(os.Args)], " "))

				// Check entrypoint is correct
				if !cmp.Equal(entrypoint, "/bin/sh") {
					fmt.Fprintf(os.Stderr, "entrypoint mismatch (-want +got):\n%s", cmp.Diff("/bin/sh", entrypoint))
					os.Exit(1)
				}

				// Check command is correct
				if !cmp.Equal(command, "command arg1") {
					fmt.Fprintf(os.Stderr, "command mismatch (-want +got):\n%s", cmp.Diff("command arg1", command))
					os.Exit(1)
				}

				// Check piped value in is correct
				if !cmp.Equal(stdin, "pipe value") {
					fmt.Fprintf(os.Stderr, "stdin mismatch (-want +got):\n%s", cmp.Diff("pipe value", stdin))
					os.Exit(1)
				}

				fmt.Fprint(os.Stdout, "test std out")
				os.Exit(0)
			}),
		},
		{
			"Failed shell command",
			errors.New("exit status 1: shell command failed"),
			"",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 100,
				Shell: &hook.Shell{
					Command:    []string{"command"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommand("failed", func(t *testing.T) {
				fmt.Fprint(os.Stderr, "shell command failed")
				os.Exit(1)
			}),
		},
		{
			"Failed shell command timeout",
			errors.New("Entrypoint '/bin/sh', command '[command]' timed out"),
			"",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 5,
				Shell: &hook.Shell{
					Command:    []string{"command"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommand("timeout", func(t *testing.T) {
				fmt.Fprint(os.Stdout, "test std out")
				time.Sleep(10 * time.Millisecond)
				os.Exit(0)
			}),
		},
		{
			"Failed shell command timeout, multiple args",
			errors.New("Entrypoint '/bin/sh', command '[command arg1]' timed out"),
			"",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 5,
				Shell: &hook.Shell{
					Command:    []string{"command", "arg1"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommand("timeout", func(t *testing.T) {
				fmt.Fprint(os.Stdout, "test std out")
				time.Sleep(10 * time.Millisecond)
				os.Exit(0)
			}),
		},
		{
			"Failed shell command fail to start",
			errors.New("exec: already started"),
			"",
			&hook.Definition{
				Type:    shell.Type,
				Timeout: 100,
				Shell: &hook.Shell{
					Command:    []string{"command"},
					Entrypoint: "/bin/sh",
				},
			},
			"pipe value",
			fakeExecCommandAndStart("fail to start", func(t *testing.T) {
				fmt.Fprint(os.Stdout, "test std out")
				os.Exit(0)
			}),
		},
	}
	code := m.Run()
	os.Exit(code)
}

func TestExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			s := &shell.Execute{test.command}
			result, err := s.ExecuteWithValue(test.definition, test.pipeValue)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf(result)
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}

			if !cmp.Equal(result, test.expected) {
				t.Errorf("stdout mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestExecute_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		command     func(name string, arg ...string) *exec.Cmd
	}{
		{
			"Return type",
			"shell",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &shell.Execute{
				Command: test.command,
			}
			result := execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
