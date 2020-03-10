package fake_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

func TestPredicter_GetIDsToRemove(t *testing.T) {
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
		predicter   fake.Predicter
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Return error",
			nil,
			errors.New("predicter error"),
			fake.Predicter{
				GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
					return nil, errors.New("predicter error")
				},
			},
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"Return IDs",
			[]int{2, 3, 6},
			nil,
			fake.Predicter{
				GetIDsToRemoveReactor: func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
					return []int{2, 3, 6}, nil
				},
			},
			&config.Model{},
			[]*stored.Evaluation{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.predicter.GetIDsToRemove(test.model, test.evaluations)
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

func TestPredicter_GetPredictionReactor(t *testing.T) {
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
		predicter   fake.Predicter
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Return error",
			0,
			errors.New("predicter error"),
			fake.Predicter{
				GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
					return 0, errors.New("predicter error")
				},
			},
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"Return IDs",
			52,
			nil,
			fake.Predicter{
				GetPredictionReactor: func(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
					return 52, nil
				},
			},
			&config.Model{},
			[]*stored.Evaluation{},
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
				t.Errorf("config mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestPredicter_GetTypeReactor(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		predicter   fake.Predicter
	}{
		{
			"Return type",
			"example type",
			fake.Predicter{
				GetTypeReactor: func() string {
					return "example type"
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := test.predicter.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("config mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
