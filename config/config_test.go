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

package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/config"
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
		configBytes []byte
	}{
		{
			"Invalid YAML",
			nil,
			errors.New("yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into config.Config"),
			[]byte("invalid-test"),
		},
		{
			"Success default config",
			&config.Config{
				DecisionType:  "maximum",
				DBPath:        "/store/predictive-horizontal-pod-autoscaler.db",
				MigrationPath: "/app/sql",
			},
			nil,
			[]byte(""),
		},
		{
			"Success custom config",
			&config.Config{
				DecisionType:  "testDecision",
				DBPath:        "testPath",
				MigrationPath: "testMigrationPath",
				Models: []*config.Model{
					&config.Model{
						Type:        "test",
						Name:        "testPrediction",
						PerInterval: 1,
						Linear: &config.Linear{
							LookAhead:    50,
							StoredValues: 10,
						},
					},
				},
			},
			nil,
			[]byte(strings.Replace(`
			decisionType: testDecision
			dbPath: testPath
			migrationPath: testMigrationPath
			models:
			- type: test
			  name: testPrediction
			  perInterval: 1
			  linear:
			    lookAhead: 50
			    storedValues: 10
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
