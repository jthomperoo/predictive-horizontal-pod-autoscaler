/*
Copyright 2021 The Predictive Horizontal Pod Autoscaler Authors.

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

package config_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/config"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultTolerance               = float64(0.1)
	defaultCPUInitializationPeriod = 300
	defaultInitialReadinessDelay   = 30
)

func TestLoadConfig(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    *config.Config
		expectedErr error
		configBytes io.Reader
	}{
		{
			"Invalid YAML",
			nil,
			errors.New(`error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Config`),
			strings.NewReader("invalid-test"),
		},
		{
			"Success default config",
			&config.Config{
				DecisionType:            "maximum",
				DBPath:                  "/store/predictive-horizontal-pod-autoscaler.db",
				MigrationPath:           "/app/sql",
				Tolerance:               defaultTolerance,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
				InitialReadinessDelay:   defaultInitialReadinessDelay,
			},
			nil,
			strings.NewReader("valid: true"),
		},
		{
			"Success custom config, YAML",
			&config.Config{
				DecisionType:            "testDecision",
				DBPath:                  "testPath",
				MigrationPath:           "testMigrationPath",
				Tolerance:               0.5,
				CPUInitializationPeriod: 25,
				InitialReadinessDelay:   321,
				Models: []*config.Model{
					{
						Type:        "test",
						Name:        "testPrediction",
						PerInterval: 1,
						Linear: &config.Linear{
							LookAhead:    50,
							StoredValues: 10,
						},
					},
				},
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: v1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: func() *int32 { i := int32(50); return &i }(),
							},
						},
					},
				},
			},
			nil,
			strings.NewReader(strings.Replace(`
			decisionType: testDecision
			dbPath: testPath
			migrationPath: testMigrationPath
			tolerance: 0.5
			cpuInitializationPeriod: 25
			initialReadinessDelay: 321
			models:
			- type: test
			  name: testPrediction
			  perInterval: 1
			  linear:
			    lookAhead: 50
			    storedValues: 10
			metrics:
			- type: Resource
			  resource:
			    name: cpu
			    target:
			      type: Utilization
			      averageUtilization: 50
			`, "\t", "", -1)),
		},
		{
			"Success custom config, JSON",
			&config.Config{
				DecisionType:            "testDecision",
				DBPath:                  "testPath",
				MigrationPath:           "testMigrationPath",
				Tolerance:               defaultTolerance,
				CPUInitializationPeriod: defaultCPUInitializationPeriod,
				InitialReadinessDelay:   defaultInitialReadinessDelay,
				Models: []*config.Model{
					{
						Type:        "test",
						Name:        "testPrediction",
						PerInterval: 1,
						Linear: &config.Linear{
							LookAhead:    50,
							StoredValues: 10,
						},
					},
				},
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: v1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: func() *int32 { i := int32(50); return &i }(),
							},
						},
					},
				},
			},
			nil,
			strings.NewReader(strings.Replace(`
			{
				"decisionType": "testDecision",
				"dbPath": "testPath",
				"migrationPath": "testMigrationPath",
				"models": [
				  {
					"type": "test",
					"name": "testPrediction",
					"perInterval": 1,
					"linear": {
					  "lookAhead": 50,
					  "storedValues": 10
					}
				  }
				],
				"metrics": [
				  {
					"type": "Resource",
					"resource": {
					  "name": "cpu",
					  "target": {
						"type": "Utilization",
						"averageUtilization": 50
					  }
					}
				  }
				]
			  }
			`, "\t", "", -1)),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := config.LoadConfig(test.configBytes)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("config mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
