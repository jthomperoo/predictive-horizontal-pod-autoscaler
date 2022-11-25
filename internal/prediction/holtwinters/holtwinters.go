/*
Copyright 2022 The Predictive Horizontal Pod Autoscaler Authors.

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

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/hook"
)

const algorithmPath = "algorithms/holt_winters/holt_winters.py"

const (
	defaultTimeout = 30000
)

// Runner defines an algorithm runner, allowing algorithms to be run
type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// Predict provides logic for using Holt Winters to make a prediction
type Predict struct {
	HookExecute hook.Executer
	Runner      AlgorithmRunner
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

type runTimeTuningFetchHookRequest struct {
	Model          jamiethompsonmev1alpha1.Model                 `json:"model"`
	ReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas `json:"replicaHistory"`
}

type runTimeTuningFetchHookResult struct {
	Alpha *float64 `json:"alpha"`
	Beta  *float64 `json:"beta"`
	Gamma *float64 `json:"gamma"`
}

// GetPrediction uses holt winters to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
	err := p.validate(model)
	if err != nil {
		return 0, err
	}

	// Statsmodels requires at least 2 * seasonal_periods to make a prediction with Holt Winters
	// https://github.com/statsmodels/statsmodels/blob/77bb1d276c7d11bc8657497b4307aa7575c3e65c/statsmodels/tsa/exponential_smoothing/initialization.py#L57-L61
	if len(replicaHistory) < 2*model.HoltWinters.SeasonalPeriods {
		return 0, nil
	}

	// Statsmodels requires at least 10 + 2 * (seasonal_periods // 2) to make a prediction with Holt Winters
	// https://github.com/statsmodels/statsmodels/blob/77bb1d276c7d11bc8657497b4307aa7575c3e65c/statsmodels/tsa/exponential_smoothing/initialization.py#L66-L71
	if len(replicaHistory) < 10+2*(model.HoltWinters.SeasonalPeriods/2) {
		return 0, nil
	}

	alpha := model.HoltWinters.Alpha
	beta := model.HoltWinters.Beta
	gamma := model.HoltWinters.Gamma

	if model.HoltWinters.RuntimeTuningFetchHook != nil {

		// Convert request into JSON string
		request, err := json.Marshal(&runTimeTuningFetchHookRequest{
			Model:          *model,
			ReplicaHistory: replicaHistory,
		})
		if err != nil {
			// Should not occur
			panic(err)
		}

		// Request runtime tuning values
		hookResult, err := p.HookExecute.ExecuteWithValue(model.HoltWinters.RuntimeTuningFetchHook, string(request))
		if err != nil {
			return 0, err
		}

		// Parse result
		var result runTimeTuningFetchHookResult
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
		return 0, errors.New("no alpha tuning value provided for Holt-Winters prediction")
	}
	if beta == nil {
		return 0, errors.New("no beta tuning value provided for Holt-Winters prediction")
	}
	if gamma == nil {
		return 0, errors.New("no gamma tuning value provided for Holt-Winters prediction")
	}

	// Collect data for historical series
	series := make([]float64, len(replicaHistory))
	for i, timestampedReplica := range replicaHistory {
		series[i] = float64(timestampedReplica.Replicas)
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

	timeout := defaultTimeout
	if model.CalculationTimeout != nil {
		timeout = *model.CalculationTimeout
	}

	value, err := p.Runner.RunAlgorithmWithValue(algorithmPath, string(parameters), timeout)
	if err != nil {
		return 0, err
	}

	prediction, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return int32(prediction), nil
}

func (p *Predict) PruneHistory(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
	err := p.validate(model)
	if err != nil {
		return nil, err
	}

	// Sort by date created
	sort.Slice(replicaHistory, func(i, j int) bool {
		return !replicaHistory[i].Time.Before(replicaHistory[j].Time)
	})

	// If there are too many stored seasons, remove the oldest ones

	// This rounds down, so if you have 7 replica data, with seasonal period of 3 and only 2 stored seasons it will
	// round the 7 / 3 (2.34) down to 2, then it will do 2 - 2 resulting in not removing any seasons
	// This is deliberate to allow full seasons to build up before pruning the old ones
	numberOfSeasonsToRemove := len(replicaHistory)/model.HoltWinters.SeasonalPeriods - model.HoltWinters.StoredSeasons
	numberOfReplicasToRemove := len(replicaHistory) - numberOfSeasonsToRemove*model.HoltWinters.SeasonalPeriods

	for i := len(replicaHistory) - 1; i >= numberOfReplicasToRemove; i-- {
		replicaHistory = append(replicaHistory[:i], replicaHistory[i+1:]...)
	}

	return replicaHistory, nil
}

// GetType returns the type of the Prediction model
func (p *Predict) GetType() string {
	return jamiethompsonmev1alpha1.TypeHoltWinters
}

func (p *Predict) validate(model *jamiethompsonmev1alpha1.Model) error {
	if model.HoltWinters == nil {
		return errors.New("no HoltWinters configuration provided for model")
	}

	if model.HoltWinters.Trend == "" {
		return errors.New("no required 'trend' value provided for model")
	}

	return nil
}
