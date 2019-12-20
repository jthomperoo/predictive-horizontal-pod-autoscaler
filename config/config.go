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

// Package config provides configuration options for the Predictive Horizontal Pod Autoscaler
package config

import (
	"gopkg.in/yaml.v2"
)

const (
	// DecisionMaximum means use the highest predicted value from the models
	DecisionMaximum = "maximum"
	// DecisionMinimum means use the lowest predicted value from the models
	DecisionMinimum = "minimum"
	// DecisionMean means use the mean average of predicted values
	DecisionMean = "mean"
)

// Config holds the configuration of the Predictive element of the PHPA
type Config struct {
	Models        []*Model `yaml:"models"`
	DecisionType  string   `yaml:"decisionType"`
	DBPath        string   `yaml:"dbPath"`
	MigrationPath string   `yaml:"migrationPath"`
}

// Model represents a prediction model to use, e.g. a linear regression
type Model struct {
	Type        string       `yaml:"type"`
	Name        string       `yaml:"name"`
	PerInterval int          `yaml:"perInterval"`
	Linear      *Linear      `yaml:"linear"`
	HoltWinters *HoltWinters `yaml:"holtWinters"`
}

// HoltWinters represents a holt-winters exponential smoothing prediction model configuration
type HoltWinters struct {
	Alpha         float64 `yaml:"alpha"`
	Beta          float64 `yaml:"beta"`
	Gamma         float64 `yaml:"gamma"`
	SeasonLength  int     `yaml:"seasonLength"`
	StoredSeasons int     `yaml:"storedSeasons"`
	Method        string  `yaml:"method"`
}

// Linear represents a linear regression prediction model configuration
type Linear struct {
	StoredValues int `yaml:"storedValues"`
	LookAhead    int `yaml:"lookAhead"`
}

// LoadConfig takes in the predictive config as a byte array and uses it to build the config, overriding default values
func LoadConfig(configEnv []byte) (*Config, error) {
	config := newDefaultConfig()
	err := yaml.Unmarshal(configEnv, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func newDefaultConfig() *Config {
	return &Config{
		DecisionType:  "maximum",
		DBPath:        "/store/predictive-horizontal-pod-autoscaler.db",
		MigrationPath: "/app/sql",
	}
}
