package fake

import (
	"github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Store (fake) provides a way to insert functionality into a Store
type Store struct {
	GetEvaluationReactor    func(model string) ([]*stored.Evaluation, error)
	AddEvaluationReactor    func(model string, evaluation *evaluate.Evaluation) error
	RemoveEvaluationReactor func(id int) error
	GetModelReactor         func(model string) (*stored.Model, error)
	UpdateModelReactor      func(model string, intervalsPassed int) error
}

// GetEvaluation calls the fake Store function
func (f *Store) GetEvaluation(model string) ([]*stored.Evaluation, error) {
	return f.GetEvaluationReactor(model)
}

// AddEvaluation calls the fake Store function
func (f *Store) AddEvaluation(model string, evaluation *evaluate.Evaluation) error {
	return f.AddEvaluationReactor(model, evaluation)
}

// RemoveEvaluation calls the fake Store function
func (f *Store) RemoveEvaluation(id int) error {
	return f.RemoveEvaluationReactor(id)
}

// GetModel calls the fake Store function
func (f *Store) GetModel(model string) (*stored.Model, error) {
	return f.GetModelReactor(model)
}

// UpdateModel calls the fake Store function
func (f *Store) UpdateModel(model string, intervalsPassed int) error {
	return f.UpdateModelReactor(model, intervalsPassed)
}
