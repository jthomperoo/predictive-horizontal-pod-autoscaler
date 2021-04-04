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

// Package hook provides standardised way to trigger hooks and provide values at different points in execution
package hook

import (
	"fmt"
)

// Definition describes a hook for passing data/triggering logic, such as through a shell command
type Definition struct {
	Type    string `json:"type"`
	Timeout int    `json:"timeout"`
	Shell   *Shell `json:"shell"`
	HTTP    *HTTP  `json:"http"`
}

// Shell describes configuration options for a shell command hook
type Shell struct {
	Command    []string `json:"command"`
	Entrypoint string   `json:"entrypoint"`
}

// HTTP describes configuration options for an HTTP request hook
type HTTP struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	SuccessCodes  []int             `json:"successCodes"`
	ParameterMode string            `json:"parameterMode"`
}

// Executer interface provides methods for executing user logic with a value passed through to it
type Executer interface {
	ExecuteWithValue(definition *Definition, value string) (string, error)
	GetType() string
}

// CombinedType is the type of the CombinedExecute; designed to link together multiple executers
// and to provide a simplified single entry point
const CombinedType = "combined"

// CombinedExecute is an executer that contains subexecuters that it will forward hook requests
// to; designed to link together multiple executers and to provide a simplified single entry point
type CombinedExecute struct {
	Executers []Executer
}

// ExecuteWithValue takes in a hook definition and a value to pass, it will look at the stored sub executers
// and decide which executer to use for the hook provided
func (e *CombinedExecute) ExecuteWithValue(definition *Definition, value string) (string, error) {
	for _, executer := range e.Executers {
		if executer.GetType() == definition.Type {
			gathered, err := executer.ExecuteWithValue(definition, value)
			if err != nil {
				return "", err
			}
			return gathered, nil
		}
	}
	return "", fmt.Errorf("Unknown execution method: '%s'", definition.Type)
}

// GetType returns the CombinedExecute type
func (e *CombinedExecute) GetType() string {
	return CombinedType
}
