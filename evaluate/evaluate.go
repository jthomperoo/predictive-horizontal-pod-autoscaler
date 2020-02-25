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

package evaluate

import (
	"database/sql"
	"math"
	"sort"

	"github.com/jthomperoo/custom-pod-autoscaler/autoscaler"
	cpaevaluate "github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	hpaevaluate "github.com/jthomperoo/horizontal-pod-autoscaler/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// PredictiveEvaluate provides a way to make predictive evaluations
type PredictiveEvaluate struct {
	HPAEvaluator hpaevaluate.Evaluater
	Store        stored.Storer
	Predicters   []prediction.Predicter
}

// GetEvaluation takes a predictive horizontal pod autoscaler configuration and gathered metrics piped in through stdin and
// evaluates using these, returning a value of how many replicas a resource should have
func (p *PredictiveEvaluate) GetEvaluation(predictiveConfig *config.Config, metrics []*metric.Metric, runType string) (*cpaevaluate.Evaluation, error) {
	evaluation, err := p.HPAEvaluator.GetEvaluation(metrics)
	if err != nil {
		return nil, err
	}

	var predictions []int32

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
		isRunType := runType == autoscaler.RunType

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
			err = p.Store.AddEvaluation(model.Name, evaluation)
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
	targetPrediction := evaluation.TargetReplicas
	switch predictiveConfig.DecisionType {
	case config.DecisionMaximum:
		max := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction > max {
				max = prediction
			}
		}
		targetPrediction = max
		break
	case config.DecisionMinimum:
		min := int32(0)
		for i, prediction := range predictions {
			if i == 0 || prediction < min {
				min = prediction
			}
		}
		targetPrediction = min
		break
	case config.DecisionMean:
		total := int32(0)
		for _, prediction := range predictions {
			total += prediction
		}
		targetPrediction = int32(math.Ceil(float64(int(total) / len(predictions))))
		break
	case config.DecisionMedian:
		halfIndex := len(predictions) / 2
		if len(predictions)%2 == 0 {
			// Even
			targetPrediction = (predictions[halfIndex-1] + predictions[halfIndex]) / 2
		} else {
			// Odd
			targetPrediction = predictions[halfIndex]
		}
	}

	// Only use predicted if the predicted value is above the current value
	if targetPrediction > evaluation.TargetReplicas {
		evaluation.TargetReplicas = targetPrediction
	}
	return evaluation, nil
}
