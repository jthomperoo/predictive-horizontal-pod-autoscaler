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
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
)

func TestEvaluater_GetEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    *evaluate.Evaluation
		expectedErr error
		evaluater   fake.Evaluater
		metrics     []*metric.Metric
	}{
		{
			"Return error",
			nil,
			errors.New("evaluater error"),
			fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error) {
					return nil, errors.New("evaluater error")
				},
			},
			[]*metric.Metric{},
		},
		{
			"Return evaluation",
			&evaluate.Evaluation{
				TargetReplicas: 5,
			},
			nil,
			fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error) {
					return &evaluate.Evaluation{
						TargetReplicas: 5,
					}, nil
				},
			},
			[]*metric.Metric{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.evaluater.GetEvaluation(test.metrics)
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
