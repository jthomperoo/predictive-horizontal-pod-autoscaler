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

package fake_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
)

func TestAlgorithm_RunAlgorithmWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description   string
		expected      string
		expectedErr   error
		run           fake.Run
		algorithmPath string
		value         string
		timeout       int
	}{
		{
			"Return error",
			"",
			errors.New("runner error"),
			fake.Run{
				RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
					return "", errors.New("runner error")
				},
			},
			"",
			"",
			10,
		},
		{
			"Return test value",
			"test",
			nil,
			fake.Run{
				RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
					return "test", nil
				},
			},
			"",
			"",
			10,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.run.RunAlgorithmWithValue(test.algorithmPath, test.value, test.timeout)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("config mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
