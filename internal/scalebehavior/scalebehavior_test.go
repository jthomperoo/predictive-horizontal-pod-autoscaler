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

func strPtr(s string) *string {
	return &s
}

func TestDecideTargetReplicasByScalingStrategy(t *testing.T) {
	var tests = []struct {
		description       string
		expected          int32
		instance          *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler
		predictedReplicas []int32
	}{
		{
			description: "No decision type, no predicted replicas, expect 0",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{},
			},
			predictedReplicas: []int32{},
		},
		{
			description: "No decision type, use default max, 3 predicted replicas",
			expected:    10,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{},
			},
			predictedReplicas: []int32{5, 10, 7},
		},
		{
			description: "Max decision type, no predicted replicas",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMaximum),
				},
			},
			predictedReplicas: []int32{},
		},
		{
			description: "Max decision type, 5 predicted replicas",
			expected:    15,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMaximum),
				},
			},
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description: "Min decision type, no predicted replicas",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMinimum),
				},
			},
			predictedReplicas: []int32{},
		},
		{
			description: "Min decision type, 5 predicted replicas",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMinimum),
				},
			},
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description: "Mean decision type, no predicted replicas",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMean),
				},
			},
			predictedReplicas: []int32{},
		},
		{
			description: "Mean decision type, 5 predicted replicas",
			expected:    7,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMean),
				},
			},
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description: "Median decision type, no predicted replicas",
			expected:    0,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMedian),
				},
			},
			predictedReplicas: []int32{},
		},
		{
			description: "Median decision type, 5 predicted replicas",
			expected:    8,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMedian),
				},
			},
			predictedReplicas: []int32{1, 10, 8, 15, 0},
		},
		{
			description: "Median decision type, 6 predicted replicas",
			expected:    7,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					DecisionType: strPtr(jamiethompsonmev1alpha1.DecisionMedian),
				},
			},
			predictedReplicas: []int32{1, 10, 8, 15, 0, 7},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.DecideTargetReplicasByScalingStrategy(test.instance, test.predictedReplicas)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func TestDecideTargetReplicasByBehavior(t *testing.T) {
	var tests = []struct {
		description             string
		expected                int32
		instance                *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler
		currentReplicas         int32
		targetReplicas          int32
		scaleDownReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas
		scaleUpReplicaHistory   []jamiethompsonmev1alpha1.TimestampedReplicas
	}{
		{
			description: "No scaleup/scaledown histories, default behavior, max 10, no scale change",
			expected:    3,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MaxReplicas: 10,
				},
			},
			currentReplicas:         3,
			targetReplicas:          3,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, max 10, scale up from 3 to 5, within scaleup limit",
			expected:    5,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MaxReplicas: 10,
				},
			},
			currentReplicas:         3,
			targetReplicas:          5,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, max 100, scale up from 3 to 50, beyond scaleup limit",
			expected:    6,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MaxReplicas: 100,
				},
			},
			currentReplicas:         3,
			targetReplicas:          50,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, max 15, scale up from 14 to 16, beyond max limit",
			expected:    15,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MaxReplicas: 15,
				},
			},
			currentReplicas:         14,
			targetReplicas:          16,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, default min, scale down from 5 to 3, within scaledown limit",
			expected:    3,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{},
			},
			currentReplicas:         5,
			targetReplicas:          3,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, min 4, scale down from 5 to 3, beyond min limit",
			expected:    4,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(4),
				},
			},
			currentReplicas:         5,
			targetReplicas:          3,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "No scaleup/scaledown histories, default behavior, min 4, scale down from 5 to 3, beyond min limit",
			expected:    4,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(4),
				},
			},
			currentReplicas:         5,
			targetReplicas:          3,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory:   []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Scaledown history, default behavior, between 1-10, scale down from 8 to 4, stabilize to 7",
			expected:    7,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(1),
					MaxReplicas: 10,
				},
			},
			currentReplicas: 8,
			targetReplicas:  4,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{},
					Replicas: 5,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 6,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 7,
				},
			},
			scaleUpReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Scaledown history, default behavior, between 1-10, scale down from 8 to 4, stabilize to 7",
			expected:    7,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(1),
					MaxReplicas: 10,
				},
			},
			currentReplicas: 8,
			targetReplicas:  4,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{},
					Replicas: 5,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 6,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 7,
				},
			},
			scaleUpReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
		},
		{
			description: "Scaleup history, custom behavior, between 1-10, scale up from 4 to 8, stabilize to 5",
			expected:    5,
			instance: &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{
				Spec: jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec{
					MinReplicas: int32Ptr(1),
					MaxReplicas: 10,
					Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
						ScaleUp: &autoscalingv2.HPAScalingRules{
							StabilizationWindowSeconds: int32Ptr(300),
							Policies: []autoscalingv2.HPAScalingPolicy{
								{
									Type:          autoscalingv2.PercentScalingPolicy,
									PeriodSeconds: 60,
									Value:         100,
								},
							},
						},
					},
				},
			},
			currentReplicas:         4,
			targetReplicas:          8,
			scaleDownReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{},
			scaleUpReplicaHistory: []jamiethompsonmev1alpha1.TimestampedReplicas{
				{
					Time:     &metav1.Time{},
					Replicas: 7,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 6,
				},
				{
					Time:     &metav1.Time{},
					Replicas: 5,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := scalebehavior.DecideTargetReplicasByBehavior(test.instance, test.currentReplicas,
				test.targetReplicas, test.scaleDownReplicaHistory, test.scaleUpReplicaHistory)
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
