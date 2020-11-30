package algorithm

import (
	cpaconfig "github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/execute"
)

const (
	entrypoint   = "python"
	shellTimeout = 10000
)

// Runner defines an algorithm runner, allowing algorithms to be run
type Runner interface {
	RunAlgorithmWithValue(algorithmPath string, value string) (string, error)
}

// Run is an implementation of an algorithm runner that uses CPA executers to run shell commands
type Run struct {
	Executer execute.Executer
}

// RunAlgorithmWithValue runs an algorithm at the path provided, passing through the value provided
func (r *Run) RunAlgorithmWithValue(algorithmPath string, value string) (string, error) {
	return r.Executer.ExecuteWithValue(&cpaconfig.Method{
		Type:    "shell",
		Timeout: shellTimeout,
		Shell: &cpaconfig.Shell{
			Entrypoint: entrypoint,
			Command:    []string{algorithmPath},
		},
	}, value)
}
