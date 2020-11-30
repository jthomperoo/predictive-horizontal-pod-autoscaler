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

package linear

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"

	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/algorithm"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
)

// Type linear is the type of the linear predicter
const Type = "Linear"

const algorithmPath = "/app/algorithms/linear_regression/linear_regression.py"

type linearRegressionParameters struct {
	LookAhead   int                  `json:"lookAhead"`
	Evaluations []*stored.Evaluation `json:"evaluations"`
}

// Config represents a linear regression prediction model configuration
type Config struct {
	StoredValues int `yaml:"storedValues"`
	LookAhead    int `yaml:"lookAhead"`
}

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct {
	Runner algorithm.Runner
}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	if model.Linear == nil {
		return 0, errors.New("No Linear configuration provided for model")
	}

	parameters, err := json.Marshal(linearRegressionParameters{
		LookAhead:   model.Linear.LookAhead,
		Evaluations: evaluations,
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

// GetIDsToRemove provides the list of stored evaluation IDs to remove, if there are too many stored values
// it will remove the oldest ones
func (p *Predict) GetIDsToRemove(model *config.Model, evaluations []*stored.Evaluation) ([]int, error) {
	if model.Linear == nil {
		return nil, errors.New("No Linear configuration provided for model")
	}

	// Sort by date created
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].Created.Before(evaluations[j].Created)
	})
	var markedForRemove []int
	// Remove any expired values
	if len(evaluations) > model.Linear.StoredValues {
		// Remove oldest to fit into requirements
		for i := 0; i < len(evaluations)-model.Linear.StoredValues; i++ {
			markedForRemove = append(markedForRemove, evaluations[i].ID)
		}
	}
	return markedForRemove, nil
}

// GetType returns the type of the Prediction model
func (p *Predict) GetType() string {
	return Type
}
