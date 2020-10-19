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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/holtwinters"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Type HoltWinters is the type of the HoltWinters predicter
const Type = "HoltWinters"

const (
	// MethodAdditive specifies a HoltWinters time series prediction using the additive method
	MethodAdditive = "additive"
	// MethodMultiplicative specifies a HoltWinters time series prediction using the multiplicative method
	MethodMultiplicative = "multiplicative"
)

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct {
	Execute execute.Executer
}

// RunTimeTuningFetchRequest defines the request value sent as part of the method to determine the runtime Holt-Winters
// values
type RunTimeTuningFetchRequest struct {
	Model       *config.Model        `json:"model"`
	Evaluations []*stored.Evaluation `json:"evaluations"`
}

// RunTimeTuningFetchResult defines the expected response from the method that specifies the runtime Holt-Winters values
type RunTimeTuningFetchResult struct {
	Alpha *float64 `json:"alpha"`
	Beta  *float64 `json:"beta"`
	Gamma *float64 `json:"gamma"`
}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	if model.HoltWinters == nil {
		return 0, errors.New("No HoltWinters configuration provided for model")
	}

	// If less than a full season of data, return zero without error
	if len(evaluations) < model.HoltWinters.SeasonLength {
		return 0, nil
	}

	alpha := model.HoltWinters.Alpha
	beta := model.HoltWinters.Beta
	gamma := model.HoltWinters.Gamma

	if model.HoltWinters.RuntimeTuningFetch != nil {

		// Convert request into JSON string
		request, err := json.Marshal(&RunTimeTuningFetchRequest{
			Model:       model,
			Evaluations: evaluations,
		})
		if err != nil {
			// Should not occur
			panic(err)
		}

		// Request runtime tuning values
		hookResult, err := p.Execute.ExecuteWithValue(model.HoltWinters.RuntimeTuningFetch, string(request))
		if err != nil {
			return 0, err
		}

		// Parse result
		var result RunTimeTuningFetchResult
		err = json.Unmarshal([]byte(hookResult), &result)
		if err != nil {
			return 0, err
		}

		if result.Alpha != nil {
			alpha = result.Alpha
		}
		if result.Beta != nil {
			beta = result.Beta
		}
		if result.Gamma != nil {
			gamma = result.Gamma
		}
	}

	if alpha == nil {
		return 0, errors.New("No alpha tuning value provided for Holt-Winters prediction")
	}
	if beta == nil {
		return 0, errors.New("No beta tuning value provided for Holt-Winters prediction")
	}
	if gamma == nil {
		return 0, errors.New("No gamma tuning value provided for Holt-Winters prediction")
	}

	// Collect data for historical series
	series := make([]float64, len(evaluations))
	for i, evaluation := range evaluations {
		series[i] = float64(evaluation.Evaluation.TargetReplicas)
	}

	var prediction []float64
	var err error

	switch model.HoltWinters.Method {
	case MethodAdditive:
		// Build prediction 1 ahead
		prediction, err = holtwinters.PredictAdditive(series, model.HoltWinters.SeasonLength, *alpha, *beta, *gamma, 1)
		if err != nil {
			return 0, err
		}
		break
	case MethodMultiplicative:
		// Build prediction 1 ahead
		prediction, err = holtwinters.PredictMultiplicative(series, model.HoltWinters.SeasonLength, *alpha, *beta, *gamma, 1)
		if err != nil {
			return 0, err
		}
		break
	default:
		return 0, fmt.Errorf("Unknown HoltWinters method '%s'", model.HoltWinters.Method)
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
