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

package evaluate_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	cpaevaluate "github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	hpaevaluate "github.com/jthomperoo/horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

func TestGetEvaluation(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description      string
		expected         *cpaevaluate.Evaluation
		expectedErr      error
		hpaEvaluator     hpaevaluate.Evaluater
		store            stored.Storer
		predicters       []prediction.Predicter
		predictiveConfig *config.Config
		metrics          []*metric.Metric
		runType          string
	}{
		{
			"Fail, fail to get evaluation",
			nil,
			errors.New(`fail to get evaluation`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return nil, errors.New("fail to get evaluation")
				},
			},
			nil,
			nil,
			nil,
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, fail to retrieve evaluations from store",
			nil,
			errors.New(`fail to get model from store`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return nil, errors.New("fail to get model from store")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type: "fake",
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, no model exists, fail to add new",
			nil,
			errors.New(`fail to add new model`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return nil, sql.ErrNoRows
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return errors.New("fail to add new model")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type: "fake",
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, no model exists, add new and fail to retrieve newly added",
			nil,
			errors.New(`fail to get added model`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			func() *fake.Store {
				timesCalled := 0
				return &fake.Store{
					GetModelReactor: func(model string) (*stored.Model, error) {
						// First time respond that no model has been found
						if timesCalled == 0 {
							timesCalled++
							return nil, sql.ErrNoRows
						}
						return nil, errors.New("fail to get added model")
					},
					UpdateModelReactor: func(model string, intervalsPassed int) error {
						return nil
					},
				}
			}(),
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type: "fake",
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler not interval run, fail to update model",
			nil,
			errors.New(`fail to update model`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 1,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return errors.New("fail to update model")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler interval run, fail to update model",
			nil,
			errors.New(`fail to update model`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return errors.New("fail to update model")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler inteval run, fail to add evaluation",
			nil,
			errors.New(`fail to add evaluation`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return errors.New("fail to add evaluation")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler inteval run, fail to get evaluations",
			nil,
			errors.New(`fail to get evaluations`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return nil, errors.New("fail to get evaluations")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler inteval run, fail to get prediction",
			nil,
			errors.New(`fail to get prediction`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 0, errors.New("fail to get prediction")
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler inteval run, fail to get IDs to remove",
			nil,
			errors.New(`fail to get IDs to remove`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 3, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return nil, errors.New("fail to get IDs to remove")
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Fail, scaler inteval run, fail to remove evaluations",
			nil,
			errors.New(`fail to remove evaluation`),
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 3,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return errors.New("fail to remove evaluation")
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 3, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
					},
				},
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, two models, pick maximum replicas, evaluation lower than prediction",
			&cpaevaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 0,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						if model.Name == "lower" {
							return 1, nil
						}
						return 3, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "lower",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "higher",
					},
				},
				DecisionType: config.DecisionMaximum,
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, two models, pick minimum replicas, evaluation lower than prediction",
			&cpaevaluate.Evaluation{
				TargetReplicas: 1,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 2,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						if model.Name == "lower" {
							return 1, nil
						}
						return 3, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "lower",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "higher",
					},
				},
				DecisionType: config.DecisionMinimum,
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, two models, pick mean replicas, evaluation lower than prediction",
			&cpaevaluate.Evaluation{
				TargetReplicas: 1,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 0,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						if model.Name == "lower" {
							return 1, nil
						}
						return 3, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "lower",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "higher",
					},
				},
				DecisionType: config.DecisionMean,
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, four models, pick median replicas",
			&cpaevaluate.Evaluation{
				TargetReplicas: 3,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 0,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						if model.Name == "a" {
							return 10, nil
						}
						if model.Name == "b" {
							return 2, nil
						}
						if model.Name == "c" {
							return 3, nil
						}
						return 9, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "a",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "b",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "c",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "d",
					},
				},
				DecisionType: config.DecisionMedian,
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, five models, pick median replicas",
			&cpaevaluate.Evaluation{
				TargetReplicas: 4,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 0,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						if model.Name == "a" {
							return 10, nil
						}
						if model.Name == "b" {
							return 2, nil
						}
						if model.Name == "c" {
							return 3, nil
						}
						if model.Name == "d" {
							return 5, nil
						}
						return 9, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "a",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "b",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "c",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "d",
					},
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "e",
					},
				},
				DecisionType: config.DecisionMedian,
			},
			nil,
			autoscaler.RunType,
		},
		{
			"Success, one model, evaluation higher than prediction",
			&cpaevaluate.Evaluation{
				TargetReplicas: 4,
			},
			nil,
			&fake.Evaluater{
				GetEvaluationReactor: func(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error) {
					return &cpaevaluate.Evaluation{
						TargetReplicas: 4,
					}, nil
				},
			},
			&fake.Store{
				GetModelReactor: func(model string) (*stored.Model, error) {
					return &stored.Model{
						ID:              2,
						IntervalsPassed: 3,
					}, nil
				},
				UpdateModelReactor: func(model string, intervalsPassed int) error {
					return nil
				},
				AddEvaluationReactor: func(model string, evaluation *cpaevaluate.Evaluation) error {
					return nil
				},
				GetEvaluationReactor: func(model string) ([]*stored.Evaluation, error) {
					return []*stored.Evaluation{}, nil
				},
				RemoveEvaluationReactor: func(id int) error {
					return nil
				},
			},
			[]prediction.Predicter{
				&fake.Predicter{
					GetTypeReactor: func() string {
						return "fake"
					},
					GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
						return 1, nil
					},
					GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
						return []int{0, 1, 2}, nil
					},
				},
			},
			&config.Config{
				Models: []*config.Model{
					&config.Model{
						Type:        "fake",
						PerInterval: 3,
						Name:        "test",
					},
				},
				DecisionType: config.DecisionMaximum,
			},
			nil,
			autoscaler.RunType,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			evaluator := &evaluate.PredictiveEvaluate{
				HPAEvaluator: test.hpaEvaluator,
				Store:        test.store,
				Predicters:   test.predicters,
			}
			result, err := evaluator.GetEvaluation(test.predictiveConfig, test.metrics, test.runType)
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
