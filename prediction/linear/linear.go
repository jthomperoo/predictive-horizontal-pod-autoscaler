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

package linear

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/stored"
	"gonum.org/v1/gonum/stat"
)

// Type linear is the type of the linear predicter
const Type = "Linear"

// Config represents a linear regression prediction model configuration
type Config struct {
	StoredValues int `yaml:"storedValues"`
	LookAhead    int `yaml:"lookAhead"`
}

// Predict provides logic for using Linear Regression to make a prediction
type Predict struct{}

// GetPrediction uses a linear regression to predict what the replica count should be based on historical evaluations
func (p *Predict) GetPrediction(model *config.Model, evaluations []*stored.Evaluation) (int32, error) {
	if model.Linear == nil {
		return 0, errors.New("No Linear configuration provided for model")
	}

	length := len(evaluations)
	lookAhead := time.Now().UTC().Add(time.Duration(model.Linear.LookAhead) * time.Millisecond)

	var data = struct {
		x []float64
		y []float64
	}{
		x: make([]float64, length),
		y: make([]float64, length),
	}

	var max float64

	// Determine latest timestamp
	for i, savedEvaluation := range evaluations {
		timestamp := float64(savedEvaluation.Created.Unix())
		if i == 0 || max < timestamp {
			max = timestamp
		}
	}

	// Build up data for linear model, in order to not deal with huge values and get rounding errors, use the difference between
	// the time being searched for and the metric recorded time in seconds
	for i, savedEvaluation := range evaluations {
		data.x[i] = float64(lookAhead.Unix() - savedEvaluation.Created.Unix())
		data.y[i] = float64(savedEvaluation.Evaluation.TargetReplicas)
	}

	// Build model
	beta, alpha := stat.LinearRegression(data.x, data.y, nil, false)
	// Make prediction using y = alpha + (beta/maximum) * x
	// Round up
	prediction := math.Ceil(alpha + (beta/max)*float64(lookAhead.Unix()))
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
