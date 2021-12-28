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

package hook_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
)

func TestCombinedExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    string
		expectedErr error
		method      *hook.Definition
		value       string
		executers   []hook.Executer
	}{
		{
			"Fail, no executers provided",
			"",
			errors.New(`Unknown execution method: 'unknown'`),
			&hook.Definition{
				Type: "unknown",
			},
			"test",
			[]hook.Executer{},
		},
		{
			"Fail, unknown execution method",
			"",
			errors.New(`Unknown execution method: 'unknown'`),
			&hook.Definition{
				Type: "unknown",
			},
			"test",
			[]hook.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "fake"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "fake", nil
					},
				},
			},
		},
		{
			"Fail, sub executer fails",
			"",
			errors.New("execute error"),
			&hook.Definition{
				Type: "test",
			},
			"test",
			[]hook.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "", errors.New("execute error")
					},
				},
			},
		},
		{
			"Successful execute, one executer",
			"test",
			nil,
			&hook.Definition{
				Type: "test",
			},
			"test",
			[]hook.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "test", nil
					},
				},
			},
		},
		{
			"Successful execute, three executers",
			"test",
			nil,
			&hook.Definition{
				Type: "test1",
			},
			"test",
			[]hook.Executer{
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test1"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "test", nil
					},
				},
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test2"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "", nil
					},
				},
				&fake.Execute{
					GetTypeReactor: func() string {
						return "test3"
					},
					ExecuteWithValueReactor: func(method *hook.Definition, value string) (string, error) {
						return "", nil
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &hook.CombinedExecute{
				Executers: test.executers,
			}
			result, err := execute.ExecuteWithValue(test.method, test.value)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestCombinedExecute_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		executers   []hook.Executer
	}{
		{
			"Return type",
			"combined",
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			execute := &hook.CombinedExecute{
				Executers: test.executers,
			}
			result := execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
