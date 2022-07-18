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

package linear

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

const (
	defaultTimeout = 30000
)

const algorithmPath = "algorithms/linear_regression/linear_regression.py"

type linearRegressionParameters struct {
	LookAhead      int                                           `json:"lookAhead"`
	ReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas `json:"replicaHistory"`
}

// Config represents a linear regression prediction model configuration
type Config struct {
	StoredValues int `yaml:"storedValues"`
	LookAhead    int `yaml:"lookAhead"`
}

// Runner defines an algorithm runner, allowing algorithms to be run
type AlgorithmRunner interface {
	RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error)
}

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct {
	Runner AlgorithmRunner
}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
	if model.Linear == nil {
		return 0, errors.New("no Linear configuration provided for model")
	}

	if len(replicaHistory) == 0 {
		return 0, errors.New("no evaluations provided for Linear regression model")
	}

	if len(replicaHistory) == 1 {
		// If only 1 evaluation is provided do not try and calculate using the linear regression model, just return
		// the target replicas from the only evaluation
		return replicaHistory[0].Replicas, nil
	}

	parameters, err := json.Marshal(linearRegressionParameters{
		LookAhead:      model.Linear.LookAhead,
		ReplicaHistory: replicaHistory,
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
	if model.Linear == nil {
		return nil, errors.New("no Linear configuration provided for model")
	}

	if len(replicaHistory) < model.Linear.HistorySize {
		return replicaHistory, nil
	}

	// Sort by date created, newest first
	sort.Slice(replicaHistory, func(i, j int) bool {
		return !replicaHistory[i].Time.Before(replicaHistory[j].Time)
	})

	// Remove oldest to fit into requirements, have to loop from the end to allow deletion without affecting indices
	for i := len(replicaHistory) - 1; i >= model.Linear.HistorySize; i-- {
		replicaHistory = append(replicaHistory[:i], replicaHistory[i+1:]...)
	}

	return replicaHistory, nil
}

// GetType returns the type of the Prediction model
func (p *Predict) GetType() string {
	return jamiethompsonmev1alpha1.TypeLinear
}
