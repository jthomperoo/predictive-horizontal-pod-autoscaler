/*
Copyright 2019 The Predictive Horizontal Pod Autoscaler Authors.

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
	cpaconfig "github.com/jthomperoo/custom-pod-autoscaler/config"
	"github.com/jthomperoo/custom-pod-autoscaler/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

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
		description string
		expected    int32
		expectedErr error
		predicter   *holtwinters.Predict
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Fail no HoltWinters configuration",
			0,
			errors.New("No HoltWinters configuration provided for model"),
			&holtwinters.Predict{},
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"Fail, fail to runtime fetch",
			0,
			errors.New("fail runtime fetch"),
			&holtwinters.Predict{
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *cpaconfig.Method, value string) (string, error) {
						return "", errors.New("fail runtime fetch")
					}
					return &execute
				}(),
			},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					RuntimeTuningFetch: &cpaconfig.Method{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonLength: 2,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail, invalid runtime fetch response",
			0,
			errors.New("invalid character 'i' looking for beginning of value"),
			&holtwinters.Predict{
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *cpaconfig.Method, value string) (string, error) {
						return "invalid json", nil
					}
					return &execute
				}(),
			},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					RuntimeTuningFetch: &cpaconfig.Method{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonLength: 2,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail no alpha value",
			0,
			errors.New("No alpha tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "invalid",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail no beta value",
			0,
			errors.New("No beta tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Alpha:        float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "invalid",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail no gamma value",
			0,
			errors.New("No gamma tuning value provided for Holt-Winters prediction"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Alpha:        float64Ptr(0.9),
					Beta:         float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "invalid",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail invalid method",
			0,
			errors.New("Unknown HoltWinters method 'invalid'"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Alpha:        float64Ptr(0.9),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "invalid",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Fail, additive, invalid parameters",
			0,
			errors.New("Invalid parameter for prediction; alpha must be between 0 and 1, is -1.000000"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					SeasonLength: 2,
					Alpha:        float64Ptr(-1.0),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Success, use fetch but no values returned, so use hardcoded fallback",
			0,
			nil,
			&holtwinters.Predict{
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *cpaconfig.Method, value string) (string, error) {
						return `{}`, nil
					}
					return &execute
				}(),
			},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					RuntimeTuningFetch: &cpaconfig.Method{
						Type:    "test",
						Timeout: 2500,
					},
					Alpha:        float64Ptr(0.9),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Success, provide all values from fetch",
			0,
			nil,
			&holtwinters.Predict{
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *cpaconfig.Method, value string) (string, error) {
						return `{"alpha":0.2, "beta":0.2, "gamma": 0.2}`, nil
					}
					return &execute
				}(),
			},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					RuntimeTuningFetch: &cpaconfig.Method{
						Type:    "test",
						Timeout: 2500,
					},
					SeasonLength: 2,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Success, provide alpha and beta values from fetch",
			0,
			nil,
			&holtwinters.Predict{
				Execute: func() *fake.Execute {
					execute := fake.Execute{}
					execute.ExecuteWithValueReactor = func(method *cpaconfig.Method, value string) (string, error) {
						return `{"alpha":0.2, "beta":0.2}`, nil
					}
					return &execute
				}(),
			},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					RuntimeTuningFetch: &cpaconfig.Method{
						Type:    "test",
						Timeout: 2500,
					},
					Gamma:        float64Ptr(0.9),
					SeasonLength: 2,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Success, additive, less than a full season",
			0,
			nil,
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Alpha:        float64Ptr(0.9),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 5,
					Method:       "additive",
				},
			},
			[]*stored.Evaluation{},
		},
		{
			"Successful, additive",
			4,
			nil,
			&holtwinters.Predict{},
			&config.Model{
				Type: linear.Type,
				HoltWinters: &config.HoltWinters{
					Alpha:         float64Ptr(0.9),
					Beta:          float64Ptr(0.9),
					Gamma:         float64Ptr(0.9),
					SeasonLength:  3,
					StoredSeasons: 3,
					Method:        "additive",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-80) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-70) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 3,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-60) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-50) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-40) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 3,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-30) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-20) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
			},
		},
		{
			"Fail, multiplicative, invalid parameters",
			0,
			errors.New("Invalid parameter for prediction; alpha must be between 0 and 1, is -1.000000"),
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					SeasonLength: 2,
					Alpha:        float64Ptr(-1.0),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					Method:       "multiplicative",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID: 1,
				},
				&stored.Evaluation{
					ID: 2,
				},
			},
		},
		{
			"Success, multiplicative, less than a full season",
			0,
			nil,
			&holtwinters.Predict{},
			&config.Model{
				HoltWinters: &config.HoltWinters{
					Alpha:        float64Ptr(0.9),
					Beta:         float64Ptr(0.9),
					Gamma:        float64Ptr(0.9),
					SeasonLength: 5,
					Method:       "multiplicative",
				},
			},
			[]*stored.Evaluation{},
		},
		{
			"Successful, multiplicative",
			4,
			nil,
			&holtwinters.Predict{},
			&config.Model{
				Type: linear.Type,
				HoltWinters: &config.HoltWinters{
					Alpha:         float64Ptr(0.9),
					Beta:          float64Ptr(0.9),
					Gamma:         float64Ptr(0.9),
					SeasonLength:  3,
					StoredSeasons: 3,
					Method:        "multiplicative",
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-80) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-70) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 3,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-60) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-50) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-40) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 3,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-30) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
				&stored.Evaluation{
					Created: time.Now().UTC().Add(time.Duration(-20) * time.Second),
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 1,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.predicter.GetPrediction(test.model, test.evaluations)
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
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Fail no HoltWinters configuration",
			nil,
			errors.New("No HoltWinters configuration provided for model"),
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"Success remove one old season",
			[]int{14, 12, 10},
			nil,
			&config.Model{
				Type: holtwinters.Type,
				HoltWinters: &config.HoltWinters{
					SeasonLength:  3,
					StoredSeasons: 2,
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID:      1,
					Created: time.Time{}.Add(time.Duration(60) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      2,
					Created: time.Time{}.Add(time.Duration(59) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      3,
					Created: time.Time{}.Add(time.Duration(58) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      4,
					Created: time.Time{}.Add(time.Duration(57) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      5,
					Created: time.Time{}.Add(time.Duration(56) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      6,
					Created: time.Time{}.Add(time.Duration(55) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      10,
					Created: time.Time{}.Add(time.Duration(54) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      12,
					Created: time.Time{}.Add(time.Duration(53) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      14,
					Created: time.Time{}.Add(time.Duration(52) * time.Millisecond),
				},
			},
		},
		{
			"Success remove two old seasons",
			[]int{1, 2},
			nil,
			&config.Model{
				Type: holtwinters.Type,
				HoltWinters: &config.HoltWinters{
					SeasonLength:  2,
					StoredSeasons: 2,
				},
			},
			[]*stored.Evaluation{
				&stored.Evaluation{
					ID:      1,
					Created: time.Time{}.Add(time.Duration(55) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      2,
					Created: time.Time{}.Add(time.Duration(56) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      3,
					Created: time.Time{}.Add(time.Duration(57) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      4,
					Created: time.Time{}.Add(time.Duration(58) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      5,
					Created: time.Time{}.Add(time.Duration(59) * time.Millisecond),
				},
				&stored.Evaluation{
					ID:      6,
					Created: time.Time{}.Add(time.Duration(60) * time.Millisecond),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &holtwinters.Predict{}
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
