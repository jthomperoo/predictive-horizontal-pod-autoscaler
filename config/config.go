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

// Package config provides configuration options for the Predictive Horizontal Pod Autoscaler
package config

import (
	"io"

	"github.com/jthomperoo/custom-pod-autoscaler/config"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// DecisionMaximum means use the highest predicted value from the models
	DecisionMaximum = "maximum"
	// DecisionMinimum means use the lowest predicted value from the models
	DecisionMinimum = "minimum"
	// DecisionMean means use the mean average of predicted values
	DecisionMean = "mean"
	// DecisionMedian means use the median average of predicted values
	DecisionMedian = "median"
)

const (
	defaultTolerance = float64(0.1)
	// 5 minute CPU initialization period
	defaultCPUInitializationPeriod = 300
	// 30 second initial readiness delay
	defaultInitialReadinessDelay = 30
)

// Config holds the configuration of the Predictive element of the PHPA
type Config struct {
	Models                  []*Model                   `json:"models"`
	Metrics                 []autoscalingv2.MetricSpec `json:"metrics"`
	DecisionType            string                     `json:"decisionType"`
	DBPath                  string                     `json:"dbPath"`
	MigrationPath           string                     `json:"migrationPath"`
	Tolerance               float64                    `json:"tolerance"`
	CPUInitializationPeriod int                        `json:"cpuInitializationPeriod"`
	InitialReadinessDelay   int                        `json:"initialReadinessDelay"`
}

// Model represents a prediction model to use, e.g. a linear regression
type Model struct {
	Type        string       `json:"type"`
	Name        string       `json:"name"`
	PerInterval int          `json:"perInterval"`
	Linear      *Linear      `json:"linear"`
	HoltWinters *HoltWinters `json:"holtWinters"`
}

// HoltWinters represents a holt-winters exponential smoothing prediction model configuration
type HoltWinters struct {
	Alpha                *float64       `json:"alpha"`
	Beta                 *float64       `json:"beta"`
	Gamma                *float64       `json:"gamma"`
	Trend                string         `json:"trend"`
	Seasonal             string         `json:"seasonal"`
	SeasonalPeriods      int            `json:"seasonalPeriods"`
	StoredSeasons        int            `json:"storedSeasons"`
	DampedTrend          *bool          `json:"dampedTrend"`
	InitializationMethod *string        `json:"initializationMethod"`
	InitialLevel         *float64       `json:"initialLevel"`
	InitialTrend         *float64       `json:"initialTrend"`
	InitialSeasonal      *float64       `json:"initialSeasonal"`
	RuntimeTuningFetch   *config.Method `json:"runtimeTuningFetch"`
}

// Linear represents a linear regression prediction model configuration
type Linear struct {
	StoredValues int `json:"storedValues"`
	LookAhead    int `json:"lookAhead"`
}

// LoadConfig takes in the predictive config as a byte array and uses it to build the config, overriding default values
func LoadConfig(configEnv io.Reader) (*Config, error) {
	config := newDefaultConfig()
	err := yaml.NewYAMLOrJSONDecoder(configEnv, 10).Decode(&config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func newDefaultConfig() *Config {
	return &Config{
		DecisionType:            "maximum",
		DBPath:                  "/store/predictive-horizontal-pod-autoscaler.db",
		MigrationPath:           "/app/sql",
		Tolerance:               defaultTolerance,
		CPUInitializationPeriod: defaultCPUInitializationPeriod,
		InitialReadinessDelay:   defaultInitialReadinessDelay,
	}
}
