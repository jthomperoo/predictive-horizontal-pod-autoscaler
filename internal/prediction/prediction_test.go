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

package prediction_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/stored"
)

func TestModelPredict_GetPrediction(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    int32
		expectedErr error
		predicters  []prediction.Predicter
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Unknown model type",
			0,
			errors.New(`unknown model type 'invalid'`),
			[]prediction.Predicter{},
			&config.Model{
				Type: "invalid",
				Linear: &config.Linear{
					LookAhead: 10,
				},
			},
			[]*stored.Evaluation{},
		},
		{
			"GetIDsToRemove fail child predictor",
			0,
			errors.New("fail to get prediction from child"),
			[]prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 0, errors.New("fail to get prediction from child")
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				}},
			[]*stored.Evaluation{},
		},
		{
			"Successful prediction, single available model",
			3,
			nil,
			[]prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 3, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				}},
			[]*stored.Evaluation{},
		},
		{
			"Successful prediction, three available models",
			5,
			nil,
			[]prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 0, errors.New("incorrect model")
					},
					GetTypeReactor: func() string {
						return "incorrect-model"
					},
				},
				&fake.Predicter{
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 0, errors.New("incorrect model")
					},
					GetTypeReactor: func() string {
						return "incorrect-model-2"
					},
				},
				&fake.Predicter{
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 5, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				},
			},
			[]*stored.Evaluation{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &prediction.ModelPredict{
				Predicters: test.predicters,
			}
			result, err := predicter.GetPrediction(test.model, test.evaluations)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestModelPredict_GetIDsToRemove(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    []int
		expectedErr error
		predicters  []prediction.Predicter
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Unknown model type",
			nil,
			errors.New(`unknown model type 'invalid'`),
			[]prediction.Predicter{},
			&config.Model{
				Type: "invalid",
				Linear: &config.Linear{
					LookAhead: 10,
				},
			},
			[]*stored.Evaluation{},
		},
		{
			"GetIDsToRemove fail child predictor",
			nil,
			errors.New("fail to get IDs to remove"),
			[]prediction.Predicter{
				&fake.Predicter{
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return nil, errors.New("fail to get IDs to remove")
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				}},
			[]*stored.Evaluation{},
		},
		{
			"Successful GetIDsToRemove, single available model",
			[]int{5},
			nil,
			[]prediction.Predicter{
				&fake.Predicter{
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{5}, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				}},
			[]*stored.Evaluation{},
		},
		{
			"Successful GetIDsToRemove, three available models",
			[]int{5},
			nil,
			[]prediction.Predicter{
				&fake.Predicter{
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{1}, nil
					},
					GetTypeReactor: func() string {
						return "incorrect-model"
					},
				},
				&fake.Predicter{
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{2}, nil
					},
					GetTypeReactor: func() string {
						return "incorrect-model-2"
					},
				},
				&fake.Predicter{
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{5}, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			&config.Model{
				Type: "test",
				Linear: &config.Linear{
					LookAhead: 10,
				},
			},
			[]*stored.Evaluation{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &prediction.ModelPredict{
				Predicters: test.predicters,
			}
			result, err := predicter.GetIDsToRemove(test.model, test.evaluations)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("remove IDs mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestModelPredict_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
	}{
		{
			"Successful get type",
			"Model",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &prediction.ModelPredict{}
			result := predicter.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("type mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
