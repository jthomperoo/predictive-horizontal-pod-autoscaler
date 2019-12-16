package fake

import (
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
)

// Evaluater (fake) provides a way to insert functionality into a Evaluater
type Evaluater struct {
	GetEvaluationReactor func(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error)
}

// GetEvaluation calls the fake Evaluater function
func (f *Evaluater) GetEvaluation(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error) {
	return f.GetEvaluationReactor(gatheredMetrics)
}
