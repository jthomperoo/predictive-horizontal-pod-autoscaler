/*
Copyright 2020 The Predictive Horizontal Pod Autoscaler Authors.

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
	"sort"
	"strconv"

	"github.com/jthomperoo/custom-pod-autoscaler/execute"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/algorithm"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Type HoltWinters is the type of the HoltWinters predicter
const Type = "HoltWinters"

const algorithmPath = "/app/algorithms/holt_winters/holt_winters.py"

const (
	// MethodAdditive specifies a HoltWinters time series prediction using the additive method
	MethodAdditive = "additive"
	// MethodMultiplicative specifies a HoltWinters time series prediction using the multiplicative method
	MethodMultiplicative = "multiplicative"
)

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct {
	Execute execute.Executer
	Runner  algorithm.Runner
}

type holtWintersParametersParameters struct {
	Series               []float64 `json:"series"`
	Alpha                float64   `json:"alpha"`
	Beta                 float64   `json:"beta"`
	Gamma                float64   `json:"gamma"`
	Trend                string    `json:"trend"`
	Seasonal             string    `json:"seasonal"`
	SeasonalPeriods      int       `json:"seasonalPeriods"`
	DampedTrend          *bool     `json:"dampedTrend,omitempty"`
	InitializationMethod *string   `json:"initializationMethod,omitempty"`
	InitialLevel         *float64  `json:"initialLevel,omitempty"`
	InitialTrend         *float64  `json:"initialTrend,omitempty"`
	InitialSeasonal      *float64  `json:"initialSeasonal,omitempty"`
}

type runTimeTuningFetchRequest struct {
	Model       *config.Model        `json:"model"`
	Evaluations []*stored.Evaluation `json:"evaluations"`
}

type runTimeTuningFetchResult struct {
	Alpha *float64 `json:"alpha"`
	Beta  *float64 `json:"beta"`
	Gamma *float64 `json:"gamma"`
}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	if model.HoltWinters == nil {
		return 0, errors.New("No HoltWinters configuration provided for model")
	}

	// Statsmodels requires at least 10 + 2 * (seasonal_periods // 2) to make a prediction with Holt Winters
	if len(evaluations) < 10+2*model.HoltWinters.SeasonalPeriods/2 {
		return 0, nil
	}

	alpha := model.HoltWinters.Alpha
	beta := model.HoltWinters.Beta
	gamma := model.HoltWinters.Gamma

	if model.HoltWinters.RuntimeTuningFetch != nil {

		// Convert request into JSON string
		request, err := json.Marshal(&runTimeTuningFetchRequest{
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
		var result runTimeTuningFetchResult
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

	parameters, err := json.Marshal(holtWintersParametersParameters{
		Series:               series,
		Alpha:                *alpha,
		Beta:                 *beta,
		Gamma:                *gamma,
		Trend:                model.HoltWinters.Trend,
		Seasonal:             model.HoltWinters.Seasonal,
		SeasonalPeriods:      model.HoltWinters.SeasonalPeriods,
		DampedTrend:          model.HoltWinters.DampedTrend,
		InitializationMethod: model.HoltWinters.InitializationMethod,
		InitialLevel:         model.HoltWinters.InitialLevel,
		InitialTrend:         model.HoltWinters.InitialTrend,
		InitialSeasonal:      model.HoltWinters.InitialSeasonal,
	})
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	value, err := p.Runner.RunAlgorithmWithValue(algorithmPath, string(parameters))
	if err != nil {
		return 0, err
	}

	prediction, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return int32(prediction), nil
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
	seasonsToRemove := len(evaluations)/model.HoltWinters.SeasonalPeriods - model.HoltWinters.StoredSeasons
	for i := 0; i < seasonsToRemove; i++ {
		for j := 0; j < model.HoltWinters.SeasonalPeriods; j++ {
			markedForRemove = append(markedForRemove, evaluations[i+j].ID)
		}
	}
	return markedForRemove, nil
}

// GetType returns the type of the Prediction model
func (p *Predict) GetType() string {
	return Type
}
