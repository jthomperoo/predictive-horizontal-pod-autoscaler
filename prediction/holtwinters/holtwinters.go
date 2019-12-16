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

package holtwinters

import (
	"errors"
	"math"
	"sort"

	"github.com/jthomperoo/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Type HoltWinters is the type of the HoltWinters predicter
const Type = "HoltWinters"

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct{}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	if model.HoltWinters == nil {
		return 0, errors.New("No HoltWinters configuration provided for model")
	}

	// If less than a full season of data, return zero without error
	if len(evaluations) < model.HoltWinters.SeasonLength {
		return 0, nil
	}

	// Collect data for historical series
	series := make([]float64, len(evaluations))
	for i, evaluation := range evaluations {
		series[i] = float64(evaluation.Evaluation.TargetReplicas)
	}

	// Build prediction 1 ahead
	prediction, err := holtwinters.Predict(series, model.HoltWinters.SeasonLength, model.HoltWinters.Alpha, model.HoltWinters.Beta, model.HoltWinters.Gamma, 1)
	if err != nil {
		return 0, err
	}

	// Return last value in prediction
	return int32(math.Ceil(prediction[len(prediction)-1])), nil
}

// GetIDsToRemove provides the list of stored evaluation IDs to remove, if there are too many stored seasons
// it will remove the oldest seasons
func (p *Predict) GetIDsToRemove(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
	if model.HoltWinters == nil {
		return nil, errors.New("No HoltWinters configuration provided for model")
	}

	// Sort by date created
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].Created.Before(evaluations[j].Created)
	})

	var markedForRemove []int

	// If there are too many stored seasons, remove the oldest ones
	seasonsToRemove := len(evaluations)/model.HoltWinters.SeasonLength - model.HoltWinters.StoredSeasons
	for i := 0; i < seasonsToRemove; i++ {
		for j := 0; j < model.HoltWinters.SeasonLength; j++ {
			markedForRemove = append(markedForRemove, evaluations[i+j].ID)
		}
	}
	return markedForRemove, nil
}

// GetType returns the type of the Prediction model
func (p *Predict) GetType() string {
	return Type
}
