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

package algorithm_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/algorithm"
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
		// As of Go 1.20 we need a temporary directory set up to send any coverage that might be generated to
		dir, err := os.MkdirTemp("", "gotest")
		if err != nil {
			panic(err)
		}
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		argsWithoutCoverage := []string{}
		for _, arg := range args {
			if arg != "-cover" {
				argsWithoutCoverage = append(argsWithoutCoverage, arg)
			}
		}
		cs = append(cs, argsWithoutCoverage...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1", fmt.Sprintf("GOCOVERDIR=%s", dir)}
		cmd.Start()
		return cmd
	}
}

func fakeExecCommand(name string, process process) command {
	processes[name] = process
	return func(command string, args ...string) *exec.Cmd {
		// As of Go 1.20 we need a temporary directory set up to send any coverage that might be generated to
		dir, err := os.MkdirTemp("", "gotest")
		if err != nil {
			panic(err)
		}
		cs := []string{"-test.run=TestShellProcess", "--", fmt.Sprintf("-process=%s", name), command}
		fmt.Println("hello world!!")
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_TEST_PROCESS=1", fmt.Sprintf("GOCOVERDIR=%s", dir)}
		fmt.Println(os.Args[0])
		return cmd
	}
}

type test struct {
	description   string
	expectedErr   error
	expected      string
	algorithmPath string
	pipeValue     string
	timeout       int
	python        *algorithm.Python
}

var tests []test

var processes map[string]process

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	processes = map[string]process{}
	tests = []test{
		{
			description:   "Successful python command",
			expectedErr:   nil,
			expected:      "test std out",
			algorithmPath: "test-algorithm.py",
			pipeValue:     "pipe value",
			timeout:       100,
			python: &algorithm.Python{
				Command: fakeExecCommand("success", func(t *testing.T) {
					stdinb, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprint(os.Stderr, err.Error())
						os.Exit(1)
					}

					stdin := string(stdinb)
					entrypoint := strings.TrimSpace(os.Args[4])
					algorithmPath := strings.TrimSpace(strings.Join(os.Args[5:len(os.Args)], " "))

					expectedAlgorithmPath := path.Join(wd, "test-algorithm.py")

					// Check entrypoint is correct
					if !cmp.Equal(entrypoint, "python") {
						fmt.Fprintf(os.Stderr, "entrypoint mismatch (-want +got):\n%s", cmp.Diff("python", entrypoint))
						os.Exit(1)
					}

					// Check command is correct
					if !cmp.Equal(algorithmPath, expectedAlgorithmPath) {
						fmt.Fprintf(os.Stderr, "algorithmPath mismatch (-want +got):\n%s", cmp.Diff(expectedAlgorithmPath, algorithmPath))
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
				Getwd: os.Getwd,
			},
		},
		{
			description:   "Failed python command",
			expectedErr:   errors.New("exit status 1: shell command failed"),
			expected:      "",
			algorithmPath: "test-algorithm.py",
			pipeValue:     "pipe value",
			timeout:       100,
			python: &algorithm.Python{
				Command: fakeExecCommand("failed", func(t *testing.T) {
					fmt.Fprint(os.Stderr, "shell command failed")
					os.Exit(1)
				}),
				Getwd: os.Getwd,
			},
		},
		{
			description:   "Failed python command timeout",
			expectedErr:   errors.New("entrypoint 'python', command 'test-algorithm.py' timed out"),
			expected:      "",
			algorithmPath: "test-algorithm.py",
			pipeValue:     "pipe value",
			timeout:       5,
			python: &algorithm.Python{
				Command: fakeExecCommand("timeout", func(t *testing.T) {
					fmt.Fprint(os.Stdout, "test std out")
					time.Sleep(10 * time.Millisecond)
					os.Exit(0)
				}),
				Getwd: os.Getwd,
			},
		},
		{
			description:   "Failed python command fail to start",
			expectedErr:   errors.New("exec: already started"),
			expected:      "",
			algorithmPath: "test-algorithm.py",
			pipeValue:     "pipe value",
			timeout:       100,
			python: &algorithm.Python{
				Command: fakeExecCommandAndStart("fail to start", func(t *testing.T) {
					fmt.Fprint(os.Stdout, "test std out")
					os.Exit(0)
				}),
				Getwd: os.Getwd,
			},
		},
		{
			description:   "Fail to get working directory",
			expectedErr:   errors.New("fail to get working directory"),
			expected:      "",
			algorithmPath: "test-algorithm.py",
			pipeValue:     "pipe value",
			timeout:       100,
			python: &algorithm.Python{
				Command: fakeExecCommandAndStart("fail to start", func(t *testing.T) {
					fmt.Fprint(os.Stdout, "test std out")
					os.Exit(0)
				}),
				Getwd: func() (dir string, err error) {
					return "", errors.New("fail to get working directory")
				},
			},
		},
	}
	code := m.Run()
	os.Exit(code)
}

func TestPython_RunAlgorithmWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.python.RunAlgorithmWithValue(test.algorithmPath, test.pipeValue, test.timeout)
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
