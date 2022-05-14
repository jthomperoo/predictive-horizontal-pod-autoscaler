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

package evaluate

import (
	"database/sql"
	"fmt"
	"sort"

	cpaconfig "github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	cpaevaluate "github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/k8shorizmetrics/metrics"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/stored"
)

type HPAEvaluator interface {
	Evaluate(gatheredMetrics []*metrics.Metric, currentReplicas int32) (int32, error)
}

// PredictiveEvaluate provides a way to make predictive evaluations
type PredictiveEvaluate struct {
	HPAEvaluator HPAEvaluator
	Store        stored.Storer
	Predicters   []prediction.Predicter
}

// GetEvaluation takes a predictive horizontal pod autoscaler configuration and gathered metrics piped in through stdin and
// evaluates using these, returning a value of how many replicas a resource should have
func (p *PredictiveEvaluate) GetEvaluation(predictiveConfig *config.Config, metrics []*metrics.Metric, currentReplicas int32, runType string) (*cpaevaluate.Evaluation, error) {
	targetReplicas, err := p.HPAEvaluator.Evaluate(metrics, currentReplicas)
	if err != nil {
		return nil, err
	}

	predictions := []int32{targetReplicas}

	// Set up predicter with available models
	predicter := prediction.ModelPredict{
		Predicters: p.Predicters,
	}

	for _, model := range predictiveConfig.Models {
		// Get model from local storage, if it doesn't exist create it
		dbModel, err := p.Store.GetModel(model.Name)
		if err == sql.ErrNoRows {
			err = p.Store.UpdateModel(model.Name, 1)
			if err != nil {
				return nil, err
			}
			dbModel, err = p.Store.GetModel(model.Name)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		isRunInterval := dbModel.IntervalsPassed >= model.PerInterval
		isRunType := runType == cpaconfig.ScalerRunType

		// If not enough intervals have passed, increment the number of intervals passed,
		// prediction will still be calculated, but current value will not be inserted/values
		// expired
		if !isRunInterval && isRunType {
			err = p.Store.UpdateModel(model.Name, dbModel.IntervalsPassed+1)
			if err != nil {
				return nil, err
			}
		}

		// Only add new values if requested during scale and on the required interval
		if isRunInterval && isRunType {
			// Reset number of passed intervals
			err = p.Store.UpdateModel(model.Name, 1)
			if err != nil {
				return nil, err
			}
			// Add new value
			err = p.Store.AddEvaluation(model.Name, &cpaevaluate.Evaluation{
				TargetReplicas: targetReplicas,
			})
			if err != nil {
				return nil, err
			}
		}

		// Get saved values
		saved, err := p.Store.GetEvaluation(model.Name)
		if err != nil {
			return nil, err
		}

		// Get prediction, add it to the slice of predictions
		prediction, err := predicter.GetPrediction(model, saved)
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction)

		// Only remove values if requested during scale and on the required interval
		if isRunInterval && isRunType {
			valuesToRemove, err := predicter.GetIDsToRemove(model, saved)
			if err != nil {
				return nil, err
			}
			for _, val := range valuesToRemove {
				err := p.Store.RemoveEvaluation(val)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Sort predictions
	sort.Slice(predictions, func(i, j int) bool { return predictions[i] < predictions[j] })

	// Decide which prediction to use
	var targetPrediction int32
	switch predictiveConfig.DecisionType {
	case config.DecisionMaximum:
		max := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction > max {
				max = prediction
			}
		}
		targetPrediction = max
	case config.DecisionMinimum:
		min := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction < min {
				min = prediction
			}
		}
		targetPrediction = min
	case config.DecisionMean:
		total := int32(0)
		for _, prediction := range predictions {
			total += prediction
		}
		targetPrediction = int32(float64(int(total) / len(predictions)))
	case config.DecisionMedian:
		halfIndex := len(predictions) / 2
		if len(predictions)%2 == 0 {
			// Even
			targetPrediction = (predictions[halfIndex-1] + predictions[halfIndex]) / 2
		} else {
			// Odd
			targetPrediction = predictions[halfIndex]
		}
	default:
		return nil, fmt.Errorf("unknown decision type '%s'", predictiveConfig.DecisionType)
	}

	return &cpaevaluate.Evaluation{
		TargetReplicas: targetPrediction,
	}, nil
}
