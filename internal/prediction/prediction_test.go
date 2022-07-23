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
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestModelPredict_GetPrediction(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expected       int32
		expectedErr    error
		predicters     []prediction.Predicter
		model          *jamiethompsonmev1alpha1.Model
		replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
	}{
		{
			description: "Unknown model type",
			expected:    0,
			expectedErr: errors.New(`unknown model type 'invalid'`),
			predicters:  []prediction.Predicter{},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "invalid",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "GetIDsToRemove fail child predictor",
			expected:    0,
			expectedErr: errors.New("fail to get prediction from child"),
			predicters: []prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *jamiethompsonmev1alpha1.Model, evaluations []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
						return 0, errors.New("fail to get prediction from child")
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "test",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				}},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Successful prediction, single available model",
			expected:    3,
			expectedErr: nil,
			predicters: []prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *jamiethompsonmev1alpha1.Model, evaluations []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
						return 3, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "test",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				}},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Successful prediction, three available models",
			expected:    5,
			expectedErr: nil,
			predicters: []prediction.Predicter{
				&fake.Predicter{
					GetPredictionReactor: func(model *jamiethompsonmev1alpha1.Model, evaluations []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
						return 0, errors.New("incorrect model")
					},
					GetTypeReactor: func() string {
						return "incorrect-model"
					},
				},
				&fake.Predicter{
					GetPredictionReactor: func(model *jamiethompsonmev1alpha1.Model, evaluations []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
						return 0, errors.New("incorrect model")
					},
					GetTypeReactor: func() string {
						return "incorrect-model-2"
					},
				},
				&fake.Predicter{
					GetPredictionReactor: func(model *jamiethompsonmev1alpha1.Model, evaluations []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
						return 5, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "test",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &prediction.ModelPredict{
				Predicters: test.predicters,
			}
			result, err := predicter.GetPrediction(test.model, test.replicaHistory)
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

func TestModelPredict_PruneHistory(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description    string
		expected       []jamiethompsonmev1alpha1.TimestampedReplicas
		expectedErr    error
		predicters     []prediction.Predicter
		model          *jamiethompsonmev1alpha1.Model
		replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
	}{
		{
			description: "Unknown model type",
			expected:    nil,
			expectedErr: errors.New(`unknown model type 'invalid'`),
			predicters:  []prediction.Predicter{},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "invalid",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Successful PruneHistory, single available model, remove last",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &v1.Time{},
					Replicas: 1,
				},
				{
					Time:     &v1.Time{},
					Replicas: 2,
				},
			},
			expectedErr: nil,
			predicters: []prediction.Predicter{
				&fake.Predicter{
					PruneHistoryReactor: func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
						replicaHistory = replicaHistory[:len(replicaHistory)-1]
						return replicaHistory, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "test",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				}},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &v1.Time{},
					Replicas: 1,
				},
				{
					Time:     &v1.Time{},
					Replicas: 2,
				},
				{
					Time:     &v1.Time{},
					Replicas: 3,
				},
			},
		},
		{
			description: "Successful PruneHistory, three available models, remove last",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &v1.Time{},
					Replicas: 1,
				},
				{
					Time:     &v1.Time{},
					Replicas: 2,
				},
			},
			expectedErr: nil,
			predicters: []prediction.Predicter{
				&fake.Predicter{
					PruneHistoryReactor: func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
						replicaHistory = replicaHistory[:len(replicaHistory)-1]
						return replicaHistory, nil
					},
					GetTypeReactor: func() string {
						return "incorrect-model"
					},
				},
				&fake.Predicter{
					PruneHistoryReactor: func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
						replicaHistory = replicaHistory[:len(replicaHistory)-1]
						return replicaHistory, nil
					},
					GetTypeReactor: func() string {
						return "incorrect-model-2"
					},
				},
				&fake.Predicter{
					PruneHistoryReactor: func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
						replicaHistory = replicaHistory[:len(replicaHistory)-1]
						return replicaHistory, nil
					},
					GetTypeReactor: func() string {
						return "test"
					},
				},
			},
			model: &jamiethompsonmev1alpha1.Model{
				Type: "test",
				Linear: &jamiethompsonmev1alpha1.Linear{
					LookAhead: 10,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &v1.Time{},
					Replicas: 1,
				},
				{
					Time:     &v1.Time{},
					Replicas: 2,
				},
				{
					Time:     &v1.Time{},
					Replicas: 3,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &prediction.ModelPredict{
				Predicters: test.predicters,
			}
			result, err := predicter.PruneHistory(test.model, test.replicaHistory)
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
			description: "Successful get type",
			expected:    "Model",
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
