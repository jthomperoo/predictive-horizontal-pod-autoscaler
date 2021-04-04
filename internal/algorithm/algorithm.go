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
	"os/exec"

	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
)

const (
	entrypoint = "python"
)

// Runner defines an algorithm runner, allowing algorithms to be run
type Runner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

type command = func(name string, arg ...string) *exec.Cmd

// Run is an implementation of an algorithm runner that uses CPA executers to run shell commands
type Run struct {
	Executer hook.Executer
}

// RunAlgorithmWithValue runs an algorithm at the path provided, passing through the value provided
func (r *Run) RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error) {
	return r.Executer.ExecuteWithValue(&hook.Definition{
		Type:    "shell",
		Timeout: timeout,
		Shell: &hook.Shell{
			Entrypoint: entrypoint,
			Command:    []string{algorithmPath},
		},
	}, value)
}
