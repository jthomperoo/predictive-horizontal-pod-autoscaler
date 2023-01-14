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

package scalebehavior_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/scalebehavior"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetLongestPolicyPeriod(t *testing.T) {
	var tests = []struct {
		description  string
		expected     int32
		scalingRules *autoscalingv2.HPAScalingRules
	}{
		{
			description:  "Nil scaling rules, expect 0",
			expected:     0,
			scalingRules: nil,
		},
		{
			description: "No scaling rules, expect 0",
			expected:    0,
			scalingRules: &autoscalingv2.HPAScalingRules{
				Policies: []autoscalingv2.HPAScalingPolicy{},
			},
		},
		{
			description: "1 scaling rule",
			expected:    1234,
			scalingRules: &autoscalingv2.HPAScalingRules{
				Policies: []autoscalingv2.HPAScalingPolicy{
					{
						Type:          autoscalingv2.PercentScalingPolicy,
						Value:         0,
						PeriodSeconds: 1234,
					},
				},
			},
		},
		{
			description: "3 scaling rules",
			expected:    4323,
			scalingRules: &autoscalingv2.HPAScalingRules{
				Policies: []autoscalingv2.HPAScalingPolicy{
					{
						Type:          autoscalingv2.PercentScalingPolicy,
						Value:         0,
						PeriodSeconds: 1234,
					},
					{
						Type:          autoscalingv2.PodsScalingPolicy,
						Value:         0,
						PeriodSeconds: 4323,
					},
					{
						Type:          autoscalingv2.PodsScalingPolicy,
						Value:         0,
						PeriodSeconds: 0,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.GetLongestPolicyPeriod(test.scalingRules)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func secondsAfterZeroTime(seconds int) time.Time {
	return time.Time{}.Add(time.Duration(seconds) * time.Second)
}

func TestPruneTimestampedReplicasToWindow(t *testing.T) {
	var tests = []struct {
		description         string
		expected            []jamiethompsonmev1alpha1.TimestampedReplicas
		timestampedReplicas []jamiethompsonmev1alpha1.TimestampedReplicas
		window              int32
		now                 time.Time
	}{
		{
			description:         "No timestamped replicas",
			expected:            []jamiethompsonmev1alpha1.TimestampedReplicas{},
			timestampedReplicas: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			window:              300,
			now:                 time.Time{},
		},
		{
			description: "1 timestamped replica in range",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(500)},
					Replicas: 30,
				},
			},
			timestampedReplicas: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(500)},
					Replicas: 30,
				},
			},
			window: 300,
			now:    secondsAfterZeroTime(600),
		},
		{
			description: "1 timestamped replica not in range",
			expected:    []jamiethompsonmev1alpha1.TimestampedReplicas{},
			timestampedReplicas: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(299)},
					Replicas: 30,
				},
			},
			window: 300,
			now:    secondsAfterZeroTime(600),
		},
		{
			description: "5 timestamped replicas, 2 not in range",
			expected: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(599)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(500)},
					Replicas: 3,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(301)},
					Replicas: 4,
				},
			},
			timestampedReplicas: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(599)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(150)}, // Not in range
					Replicas: 2,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(500)},
					Replicas: 3,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(301)},
					Replicas: 4,
				},
				{
					Time:     &metav1.Time{Time: secondsAfterZeroTime(280)}, // Not in range
					Replicas: 5,
				},
			},
			window: 300,
			now:    secondsAfterZeroTime(600),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.PruneTimestampedReplicasToWindow(test.timestampedReplicas, test.window, test.now)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestDecideTargetReplicasByScalingStrategy(t *testing.T) {
	var tests = []struct {
		description       string
		expected          int32
		decisionType      string
		predictedReplicas []int32
	}{
		{
			description:       "Max decision type, no predicted replicas",
			expected:          0,
			decisionType:      jamiethompsonmev1alpha1.DecisionMaximum,
			predictedReplicas: []int32{},
		},
		{
			description:       "Max decision type, 5 predicted replicas",
			expected:          15,
			decisionType:      jamiethompsonmev1alpha1.DecisionMaximum,
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description:       "Min decision type, no predicted replicas",
			expected:          0,
			decisionType:      jamiethompsonmev1alpha1.DecisionMinimum,
			predictedReplicas: []int32{},
		},
		{
			description:       "Min decision type, 5 predicted replicas",
			expected:          0,
			decisionType:      jamiethompsonmev1alpha1.DecisionMinimum,
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description:       "Mean decision type, no predicted replicas",
			expected:          0,
			decisionType:      jamiethompsonmev1alpha1.DecisionMean,
			predictedReplicas: []int32{},
		},
		{
			description:       "Mean decision type, 5 predicted replicas",
			expected:          7,
			decisionType:      jamiethompsonmev1alpha1.DecisionMean,
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description:       "Median decision type, no predicted replicas",
			expected:          0,
			decisionType:      jamiethompsonmev1alpha1.DecisionMedian,
			predictedReplicas: []int32{},
		},
		{
			description:       "Median decision type, 5 predicted replicas",
			expected:          8,
			decisionType:      jamiethompsonmev1alpha1.DecisionMedian,
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description:       "Median decision type, 6 predicted replicas",
			expected:          7,
			decisionType:      jamiethompsonmev1alpha1.DecisionMedian,
			predictedReplicas: []int32{1, 10, 8, 15, 0, 7},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.DecideTargetReplicasByScalingStrategy(test.decisionType, test.predictedReplicas)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestDecideTargetReplicasByBehavior(t *testing.T) {
	var tests = []struct {
		description             string
		expected                int32
		behavior                *autoscalingv2.HorizontalPodAutoscalerBehavior
		currentReplicas         int32
		targetReplicas          int32
		minReplicas             int32
		maxReplicas             int32
		scaleUpReplicaHistory   []jamiethompsonmev1alpha1.TimestampedReplicas
		scaleDownReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
		scaleUpEventHistory     []jamiethompsonmev1alpha1.TimestampedReplicas
		scaleDownEventHistory   []jamiethompsonmev1alpha1.TimestampedReplicas
	}{}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.DecideTargetReplicasByBehavior(test.behavior, test.currentReplicas,
				test.minReplicas, test.maxReplicas, test.targetReplicas, test.scaleUpEventHistory,
				test.scaleDownReplicaHistory, test.scaleUpEventHistory, test.scaleDownEventHistory)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
