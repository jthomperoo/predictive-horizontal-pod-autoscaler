package fake

import (
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Predicter (fake) provides a way to insert functionality into a Predicter
type Predicter struct {
	GetPredictionReactor  func(model *config.Model, evaluations []*stored.Evaluation) (int32, error)
	GetIDsToRemoveReactor func(model *config.Model, evaluations []*stored.Evaluation) ([]int, error)
	GetTypeReactor        func() string
}

// GetIDsToRemove calls the fake Predicter function
func (f *Predicter) GetIDsToRemove(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
	return f.GetIDsToRemoveReactor(model, evaluations)
}

// GetPrediction calls the fake Predicter function
func (f *Predicter) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	return f.GetPredictionReactor(model, evaluations)
}

// GetType calls the fake Predicter function
func (f *Predicter) GetType() string {
	return f.GetTypeReactor()
}
