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
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/stored"
)

func TestStore_GetEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    []*stored.Evaluation
		expectedErr error
		store       fake.Store
		model       string
	}{
		{
			"Return error",
			nil,
			errors.New("store error"),
			fake.Store{
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return nil, errors.New("store error")
				},
			},
			"test",
		},
		{
			"Return evaluation",
			[]*stored.Evaluation{
				{
					ID: 3,
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 2,
					},
				},
			},
			nil,
			fake.Store{
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{
						{
							ID: 3,
							Evaluation: stored.DBEvaluation{
								TargetReplicas: 2,
							},
						},
					}, nil
				},
			},
			"test",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.store.GetEvaluation(test.model)
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

func TestStore_AddEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expectedErr error
		store       fake.Store
		model       string
		evaluation  *evaluate.Evaluation
	}{
		{
			"Return error",
			errors.New("store error"),
			fake.Store{
				AddEvaluationReactor: func(model string, evaluation *evaluate.Evaluation) error {
					return errors.New("store error")
				},
			},
			"test",
			&evaluate.Evaluation{},
		},
		{
			"Return no error",
			nil,
			fake.Store{
				AddEvaluationReactor: func(model string, evaluation *evaluate.Evaluation) error {
					return nil
				},
			},
			"test",
			&evaluate.Evaluation{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.store.AddEvaluation(test.model, test.evaluation)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
		})
	}
}

func TestStore_RemoveEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expectedErr error
		store       fake.Store
		id          int
	}{
		{
			"Return error",
			errors.New("store error"),
			fake.Store{
				RemoveEvaluationReactor: func(id int) error {
					return errors.New("store error")
				},
			},
			1,
		},
		{
			"Return no error",
			nil,
			fake.Store{
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			3,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.store.RemoveEvaluation(test.id)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
		})
	}
}

func TestStore_GetModelReactor(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    *stored.Model
		expectedErr error
		store       fake.Store
		model       string
	}{
		{
			"Return error",
			nil,
			errors.New("store error"),
			fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return nil, errors.New("store error")
				},
			},
			"test",
		},
		{
			"Return model",
			&stored.Model{
				ID:              3,
				IntervalsPassed: 2,
				Name:            "test",
			},
			nil,
			fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              3,
						IntervalsPassed: 2,
						Name:            "test",
					}, nil
				},
			},
			"test",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.store.GetModel(test.model)
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

func TestStore_UpdateModelReactor(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description     string
		expectedErr     error
		store           fake.Store
		model           string
		intervalsPassed int
	}{
		{
			"Return error",
			errors.New("store error"),
			fake.Store{
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return errors.New("store error")
				},
			},
			"test",
			1,
		},
		{
			"Return no error",
			nil,
			fake.Store{
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
			},
			"test",
			3,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := test.store.UpdateModel(test.model, test.intervalsPassed)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
		})
	}
}
