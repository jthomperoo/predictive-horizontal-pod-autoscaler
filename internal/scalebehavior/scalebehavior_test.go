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
		now                     time.Time
	}{
		{
			description:             "Scale 2 -> 4, no scaling history or events, default behavior, min 1, max 10, scale to 4",
			expected:                4,
			behavior:                defaultBehavior(),
			currentReplicas:         2,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:             "Scale 2 -> 4, no scaling history or events, default behavior, min 1, max 3, scale to 3",
			expected:                3,
			behavior:                defaultBehavior(),
			currentReplicas:         2,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             3,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:             "Scale 2 -> 1, no scaling history or events, default behavior, min 1, max 4, scale to 1",
			expected:                1,
			behavior:                defaultBehavior(),
			currentReplicas:         2,
			targetReplicas:          1,
			minReplicas:             1,
			maxReplicas:             4,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:             "Scale 2 -> 0, no scaling history or events, default behavior, min 1, max 4, scale to 1",
			expected:                1,
			behavior:                defaultBehavior(),
			currentReplicas:         2,
			targetReplicas:          0,
			minReplicas:             1,
			maxReplicas:             4,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:             "Scale 2 -> 0, no scaling history or events, default behavior, min 0, max 4, scale to 0",
			expected:                0,
			behavior:                defaultBehavior(),
			currentReplicas:         2,
			targetReplicas:          0,
			minReplicas:             0,
			maxReplicas:             4,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:           "Scale 9 -> 2, downscale history max 8, no scaling events, default behavior, min 1, max 10, apply downscale stabilization, scale to 9",
			expected:              9,
			behavior:              defaultBehavior(),
			currentReplicas:       9,
			targetReplicas:        2,
			minReplicas:           1,
			maxReplicas:           10,
			scaleUpReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 7,
				},
				{
					Replicas: 6,
				},
				{
					Replicas: 9,
				},
				{
					Replicas: 8,
				},
			},
			scaleUpEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{},
		},
		{
			description: "Scale 3 -> 7, upscale history min 3, no scaling events, upscale stablization window 60, min 1, max 10, apply upscale stabilization, scale to 3",
			expected:    3,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: int32Ptr(60),
				},
			}),
			currentReplicas: 3,
			targetReplicas:  7,
			minReplicas:     1,
			maxReplicas:     10,
			scaleUpReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Replicas: 2,
				},
				{
					Replicas: 1,
				},
				{
					Replicas: 3,
				},
			},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description: "Scale 3 -> 7, no scaling history or events, upscaling disabled, min 1, max 10, don't scale",
			expected:    3,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.DisabledPolicySelect),
				},
			}),
			currentReplicas:         3,
			targetReplicas:          7,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description: "Scale 7 -> 3, no scaling history or events, downscaling disabled, min 1, max 10, don't scale",
			expected:    7,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.DisabledPolicySelect),
				},
			}),
			currentReplicas:         7,
			targetReplicas:          3,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                     time.Time{},
		},
		{
			description:             "Scale 4 -> 6, no scaling history, 4 pods already added last min, default behavior, min 1, max 10, apply pod policy, scale to 4",
			expected:                4,
			behavior:                defaultBehavior(),
			currentReplicas:         4,
			targetReplicas:          6,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(16 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(31 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(46 * time.Second)},
					Replicas: 1,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
		{
			description:             "Scale 4 -> 8, no scaling history, 4 pods added in last minute, 2 removed, default behavior, min 1, max 10, apply pod policy, scale to 6",
			expected:                6,
			behavior:                defaultBehavior(),
			currentReplicas:         4,
			targetReplicas:          6,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(11 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(21 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(31 * time.Second)},
					Replicas: 1,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(41 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(51 * time.Second)},
					Replicas: 1,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 4 -> 7, no scaling history, 1 pods already added last min, default behavior with min select policy, min 1, max 10, apply percent policy, scale to 6",
			expected:    6,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.MinChangePolicySelect),
				},
			}),
			currentReplicas:         4,
			targetReplicas:          7,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
		{
			description:             "Scale 30 -> 50, no scaling history, scale up history start at 20, default behavior, min 1, max 100, apply percent policy, scale to 40",
			expected:                40,
			behavior:                defaultBehavior(),
			currentReplicas:         30,
			targetReplicas:          50,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 10,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
		{
			description:             "Scale 6 -> 4, no scaling history, scale down history start at 10, default behavior, min 1, max 10, apply percent policy (no minimum), scale to 4",
			expected:                4,
			behavior:                defaultBehavior(),
			currentReplicas:         6,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(16 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(31 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(46 * time.Second)},
					Replicas: 1,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 6 -> 4, no scaling history, 4 pods already removed, only allow removing 5 pods a minute, min 1, max 10, apply pod policy, scale to 5",
			expected:    5,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         5,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         6,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(16 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(31 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(46 * time.Second)},
					Replicas: 1,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 6 -> 3, no scaling history, 4 pods already removed, 2 added, only allow removing 5 pods a minute, min 1, max 10, apply pod policy, scale to 4",
			expected:    4,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         5,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         6,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(41 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(51 * time.Second)},
					Replicas: 1,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(11 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(21 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(31 * time.Second)},
					Replicas: 1,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 8 -> 4, no scaling history, scale down history start at 10, only allow removing half pods every min, min 1, max 10, apply percent policy, scale to 5",
			expected:    5,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         50,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         8,
			targetReplicas:          4,
			minReplicas:             1,
			maxReplicas:             10,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(16 * time.Second)},
					Replicas: 1,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 80 -> 40, no scaling history, scale down history start at 100, only allow removing half pods every min, only allow 10 pods to be removed every min, select policy max, min 1, max 100, apply percent policy, scale to 50",
			expected:    50,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         50,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         10,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         80,
			targetReplicas:          40,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 10,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 10,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 80 -> 40, no scaling history, scale down history start at 100, only allow removing half pods every min, only allow 10 pods to be removed every min, select policy min, min 1, max 100, apply pod policy, scale to 80",
			expected:    80,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleDown: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.MinChangePolicySelect),
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         50,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         10,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         80,
			targetReplicas:          40,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory:     []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 10,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 10,
				},
			},
			now: time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 40 -> 80, no scaling history, scale up history start at 30, only allow adding three times as many pods every min, only allow 20 pods to be added every min, select policy min, min 1, max 100, apply pod policy, scale to 50",
			expected:    50,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.MinChangePolicySelect),
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         150,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         20,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         40,
			targetReplicas:          80,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 5,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(2 * time.Second)},
					Replicas: 5,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 40 -> 100, no scaling history, scale up history start at 30, only allow adding three times as many pods every min, only allow 20 pods to be added every min, select policy max, min 1, max 100, apply pod policy, scale to 90",
			expected:    90,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         200,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         20,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         40,
			targetReplicas:          100,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 5,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(2 * time.Second)},
					Replicas: 5,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
		{
			description: "Scale 3 -> 6, no scaling history, scale up history start at 0, only allow adding three times as many pods every min, only allow 5 pods to be added every min, select policy min, min 0, max 100, apply percent policy, don't scale",
			expected:    3,
			behavior: mergeWithDefault(&autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					SelectPolicy: selectPolicyPtr(autoscalingv2.MinChangePolicySelect),
					Policies: []autoscalingv2.HPAScalingPolicy{
						{
							Type:          autoscalingv2.PercentScalingPolicy,
							Value:         200,
							PeriodSeconds: 60,
						},
						{
							Type:          autoscalingv2.PodsScalingPolicy,
							Value:         5,
							PeriodSeconds: 60,
						},
					},
				},
			}),
			currentReplicas:         3,
			targetReplicas:          6,
			minReplicas:             1,
			maxReplicas:             100,
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(1 * time.Second)},
					Replicas: 1,
				},
				{
					Time:     &metav1.Time{Time: time.Time{}.Add(2 * time.Second)},
					Replicas: 2,
				},
			},
			scaleDownEventHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			now:                   time.Time{}.Add(60 * time.Second),
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.DecideTargetReplicasByBehavior(test.behavior, test.currentReplicas,
				test.targetReplicas, test.minReplicas, test.maxReplicas, test.scaleUpReplicaHistory,
				test.scaleDownReplicaHistory, test.scaleUpEventHistory, test.scaleDownEventHistory, test.now)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

// Downscale constants
const (
	defaultDownscaleStabilization                 = int32(300)
	defaultDownscalePercentagePolicyPeriodSeconds = int32(60)
	defaultDownscalePercentagePolicyValue         = int32(100)
)

// Upscale constants
const (
	defaultUpscaleStabilization                 = int32(0)
	defaultUpscalePercentagePolicyPeriodSeconds = int32(60)
	defaultUpscalePercentagePolicyValue         = int32(100)
	defaultUpscalePodsPolicyPeriodSeconds       = int32(60)
	defaultUpscalePodsPolicyValue               = int32(4)
)

func mergeWithDefault(behavior *autoscalingv2.HorizontalPodAutoscalerBehavior) *autoscalingv2.HorizontalPodAutoscalerBehavior {
	if behavior == nil {
		return &autoscalingv2.HorizontalPodAutoscalerBehavior{
			ScaleDown: defaultDownscale(),
			ScaleUp:   defaultUpscale(),
		}
	}

	// We need to take a deep copy here, since we don't want any defaults we fill in to be persisted on the
	// actual object
	behavior = behavior.DeepCopy()

	behavior.ScaleDown = copyHPAScalingRules(behavior.ScaleDown, defaultDownscale())
	behavior.ScaleUp = copyHPAScalingRules(behavior.ScaleUp, defaultUpscale())

	return behavior
}

func copyHPAScalingRules(from, to *autoscalingv2.HPAScalingRules) *autoscalingv2.HPAScalingRules {
	if from == nil {
		return to
	}
	if from.SelectPolicy != nil {
		to.SelectPolicy = from.SelectPolicy
	}
	if from.StabilizationWindowSeconds != nil {
		to.StabilizationWindowSeconds = from.StabilizationWindowSeconds
	}
	if from.Policies != nil {
		to.Policies = from.Policies
	}
	return to
}

func defaultBehavior() *autoscalingv2.HorizontalPodAutoscalerBehavior {
	return &autoscalingv2.HorizontalPodAutoscalerBehavior{
		ScaleDown: defaultDownscale(),
		ScaleUp:   defaultUpscale(),
	}
}

func defaultDownscale() *autoscalingv2.HPAScalingRules {
	return &autoscalingv2.HPAScalingRules{
		StabilizationWindowSeconds: int32Ptr(defaultDownscaleStabilization),
		SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
		Policies: []autoscalingv2.HPAScalingPolicy{
			{
				Type:          autoscalingv2.PercentScalingPolicy,
				PeriodSeconds: defaultDownscalePercentagePolicyPeriodSeconds,
				Value:         defaultDownscalePercentagePolicyValue,
			},
		},
	}
}

func defaultUpscale() *autoscalingv2.HPAScalingRules {
	return &autoscalingv2.HPAScalingRules{
		StabilizationWindowSeconds: int32Ptr(0),
		SelectPolicy:               selectPolicyPtr(autoscalingv2.MaxChangePolicySelect),
		Policies: []autoscalingv2.HPAScalingPolicy{
			{
				Type:          autoscalingv2.PercentScalingPolicy,
				PeriodSeconds: defaultUpscalePercentagePolicyPeriodSeconds,
				Value:         defaultUpscalePercentagePolicyValue,
			},
			{
				Type:          autoscalingv2.PodsScalingPolicy,
				PeriodSeconds: defaultUpscalePodsPolicyPeriodSeconds,
				Value:         defaultUpscalePodsPolicyValue,
			},
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func selectPolicyPtr(policy autoscalingv2.ScalingPolicySelect) *autoscalingv2.ScalingPolicySelect {
	return &policy
}
