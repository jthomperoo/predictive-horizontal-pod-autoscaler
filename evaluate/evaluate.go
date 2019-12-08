package evaluate

import (
	"database/sql"
	"math"

	cpaevaluate "github.com/jthomperoo/custom-pod-autoscaler/evaluate"
	"github.com/jthomperoo/custom-pod-autoscaler/scaler"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// HPAEvaluator represents a way to interact with the horizontal-pod-autoscaler evaluation logic
type HPAEvaluator interface {
	GetEvaluation(gatheredMetrics []*metric.Metric) (*cpaevaluate.Evaluation, error)
}

// PredictiveEvaluate provides a way to make predictive evaluations
type PredictiveEvaluate struct {
	HPAEvaluator HPAEvaluator
	Store        stored.Store
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
		Predicters: []prediction.Predicter{
			&linear.Predict{},
		},
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
		}
		if err != nil {
			return nil, err
		}

		updateValues := dbModel.IntervalsPassed < model.PerInterval && runType == scaler.RunType

		// If not enough intervals have passed, increment the number of intervals passed,
		// prediction will still be calculated, but current value will not be inserted/values
		// expired
		if updateValues {
			err = p.Store.UpdateModel(model.Name, dbModel.IntervalsPassed+1)
			if err != nil {
				return nil, err
			}
		}

		// Only add new values if requested during scale and on the required interval
		if updateValues {
			// Add new value
			err = p.Store.Add(model.Name, evaluation)
			if err != nil {
				return nil, err
			}
		}

		// Get saved values
		saved, err := p.Store.Get(model.Name)
		if err != nil {
			return nil, err
		}

		prediction, err := predicter.GetPrediction(model, saved)
		if err != nil {
			return nil, err
		}

		predictions = append(predictions, prediction)

		// Only remove values if requested during scale and on the required interval
		if updateValues {
			valuesToRemove, err := predicter.GetIDsToRemove(model, saved)
			if err != nil {
				return nil, err
			}
			for _, val := range valuesToRemove {
				p.Store.Remove(val)
			}
		}
	}

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
		targetPrediction = int32(math.Ceil(float64(int(total) / len(predictions))))
	}

	// Only use predicted if the predicted value is above the current value
	if targetPrediction > evaluation.TargetReplicas {
		evaluation.TargetReplicas = targetPrediction
	}
	return evaluation, nil
}
