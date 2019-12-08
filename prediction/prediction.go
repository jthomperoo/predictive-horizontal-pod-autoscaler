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

// Package prediction provides a framework for using models to make predictions based on historical evaluations
package prediction

import (
	"fmt"

	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Predicter is an interface providing methods for making a prediction based on a model, a time to predict and values
type Predicter interface {
	GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error)
	GetIDsToRemove(model *config.Model, evaluations []*stored.Evaluation) ([]int, error)
	GetType() string
}

// ModelPredict is used to route a prediction to the appropriate predicter based on the model provided
// Should be initialised with available predicters for it to use
type ModelPredict struct {
	Predicters []Predicter
}

// GetPrediction generates a prediction for any model that the ModelPredict has been set up to use
func (m *ModelPredict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	for _, predicter := range m.Predicters {
		if predicter.GetType() == model.Type {
			return predicter.GetPrediction(model, evaluations)
		}
	}
	return 0, fmt.Errorf("Unknown model type '%s'", model.Type)
}

// GetIDsToRemove finds the appropriate logic for the model and gets a list of stored IDs to remove
func (m *ModelPredict) GetIDsToRemove(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
	for _, predicter := range m.Predicters {
		if predicter.GetType() == model.Type {
			return predicter.GetIDsToRemove(model, evaluations)
		}
	}
	return nil, fmt.Errorf("Unknown model type '%s'", model.Type)
}

// GetType returns the type of the ModelPredict, "Model"
func (m *ModelPredict) GetType() string {
	return "Model"
}
