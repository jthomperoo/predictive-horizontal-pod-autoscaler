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

package holtwinters_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/holtwinters"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func intPtr(i int) *int {
	return &i
}

func float64Ptr(val float64) *float64 {
	return &val
}

func TestPredict_GetPrediction(t *testing.T) {
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
		predicter      *holtwinters.Predict
		model          *jamiethompsonmev1alpha1.Model
		replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
	}{
		{
			"Fail no HoltWinters configuration",
			0,
			errors.New("no HoltWinters configuration provided for model"),
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			"Fail no trend configuration",
			0,
			errors.New("no required 'trend' value provided for model"),
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				Type: jamiethompsonmev1alpha1.TypeLinear,
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Trend: "",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			"Success, less than 2 * seasonal_periods observations",
			0,
			nil,
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				Type: jamiethompsonmev1alpha1.TypeLinear,
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 3,
					StoredSeasons:   3,
					Trend:           "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-80) * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-70) * time.Second)},
					Replicas: 3,
				},
			},
		},
		{
			"Success, less than 10 + 2 * (seasonal_periods // 2) observations",
			0,
			nil,
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				Type: jamiethompsonmev1alpha1.TypeLinear,
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 3,
					StoredSeasons:   3,
					Trend:           "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-80) * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-70) * time.Second)},
					Replicas: 3,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-60) * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-50) * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-40) * time.Second)},
					Replicas: 3,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-30) * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Now().UTC().Add(time.Duration(-20) * time.Second)},
					Replicas: 1,
				},
			},
		},
		{
			"Fail, fail to runtime fetch",
			0,
			errors.New("fail runtime fetch"),
			&holtwinters.Predict{
				HookExecute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
						return "", errors.New("fail runtime fetch")
					}
					return &execute
				}(),
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					RuntimeTuningFetchHook: &jamiethompsonmev1alpha1.HookDefinition{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail, invalid runtime fetch response",
			0,
			errors.New("invalid character 'i' looking for beginning of value"),
			&holtwinters.Predict{
				HookExecute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
						return "invalid json", nil
					}
					return &execute
				}(),
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					RuntimeTuningFetchHook: &jamiethompsonmev1alpha1.HookDefinition{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail no alpha value",
			0,
			errors.New("no alpha tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail no beta value",
			0,
			errors.New("no beta tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail no gamma value",
			0,
			errors.New("no gamma tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail, additive, fail to run holt winters algorithm",
			0,
			errors.New("holt winters algorithm error"),
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "", errors.New("holt winters algorithm error")
					},
				},
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "additive",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Fail, additive, holt winters algorithm invalid response",
			0,
			errors.New(`strconv.Atoi: parsing "invalid": invalid syntax`),
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "invalid", nil
					},
				},
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "additive",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Success",
			0,
			nil,
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "0", nil
					},
				},
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Success, configure calculation timeout",
			0,
			nil,
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "0", nil
					},
				},
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
				CalculationTimeout: intPtr(10),
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Success, use fetch but no values returned, so use hardcoded fallback",
			0,
			nil,
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "0", nil
					},
				},
				HookExecute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
						return `{}`, nil
					}
					return &execute
				}(),
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					RuntimeTuningFetchHook: &jamiethompsonmev1alpha1.HookDefinition{
						Type:    "test",
						Timeout: 2500,
					},
					Alpha:           float64Ptr(0.9),
					Beta:            float64Ptr(0.9),
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Success, provide all values from fetch",
			2,
			nil,
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "2", nil
					},
				},
				HookExecute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
						return `{"alpha":0.2, "beta":0.2, "gamma": 0.2}`, nil
					}
					return &execute
				}(),
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					RuntimeTuningFetchHook: &jamiethompsonmev1alpha1.HookDefinition{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
		{
			"Success, provide alpha and beta values from fetch",
			3,
			nil,
			&holtwinters.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "3", nil
					},
				},
				HookExecute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(definition *jamiethompsonmev1alpha1.HookDefinition, value string) (string, error) {
						return `{"alpha":0.2, "beta":0.2}`, nil
					}
					return &execute
				}(),
			},
			&jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					RuntimeTuningFetchHook: &jamiethompsonmev1alpha1.HookDefinition{
						Type:    "test",
						Timeout: 2500,
					},
					Gamma:           float64Ptr(0.9),
					SeasonalPeriods: 2,
					Trend:           "add",
					Seasonal:        "add",
				},
			},
			[]jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 1,
				},
				{
					Replicas: 2,
				},
				{
					Replicas: 3,
				},
				{
					Replicas: 4,
				},
				{
					Replicas: 5,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 7,
				},
				{
					Replicas: 8,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 10,
				},
				{
					Replicas: 11,
				},
				{
					Replicas: 12,
				},
				{
					Replicas: 13,
				},
				{
					Replicas: 14,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.predicter.GetPrediction(test.model, test.replicaHistory)
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
		model          *jamiethompsonmev1alpha1.Model
		replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
	}{
		{
			description:    "Fail no HoltWinters configuration",
			expected:       nil,
			expectedErr:    errors.New("no HoltWinters configuration provided for model"),
			model:          &jamiethompsonmev1alpha1.Model{},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Fail no trend configuration",
			expected:    nil,
			expectedErr: errors.New("no required 'trend' value provided for model"),
			model: &jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "6 in history, seasonal period 2, 3 stored seasons, don't prune",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
			},
			expectedErr: nil,
			model: &jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Trend:           "add",
					StoredSeasons:   3,
					SeasonalPeriods: 2,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
			},
		},
		{
			description: "7 in history, seasonal period 2, 3 stored seasons, don't prune",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
			},
			expectedErr: nil,
			model: &jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Trend:           "add",
					StoredSeasons:   3,
					SeasonalPeriods: 2,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
			},
		},
		{
			description: "8 in history, seasonal period 2, 3 stored seasons, prune oldest season",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 8,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(8) * time.Second)},
				},
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
			},
			expectedErr: nil,
			model: &jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Trend:           "add",
					StoredSeasons:   3,
					SeasonalPeriods: 2,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 8,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(8) * time.Second)},
				},
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
			},
		},
		{
			description: "8 in history, unsorted, seasonal period 2, 3 stored seasons, prune oldest season",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 8,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(8) * time.Second)},
				},
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
			},
			expectedErr: nil,
			model: &jamiethompsonmev1alpha1.Model{
				HoltWinters: &jamiethompsonmev1alpha1.HoltWinters{
					Trend:           "add",
					StoredSeasons:   3,
					SeasonalPeriods: 2,
				},
			},
			replicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 6,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(6) * time.Second)},
				},
				{
					Replicas: 1,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(1) * time.Second)},
				},
				{
					Replicas: 7,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(7) * time.Second)},
				},
				{
					Replicas: 2,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(2) * time.Second)},
				},
				{
					Replicas: 5,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(5) * time.Second)},
				},
				{
					Replicas: 4,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(4) * time.Second)},
				},
				{
					Replicas: 3,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(3) * time.Second)},
				},
				{
					Replicas: 8,
					Time:     &metav1.Time{Time: time.Time{}.Add(time.Duration(8) * time.Second)},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &holtwinters.Predict{}
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

func TestPredict_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
	}{
		{
			"Successful get type",
			"HoltWinters",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := holtwinters.Predict{}
			result := predicter.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("type mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
