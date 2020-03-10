package fake_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/fake"
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
