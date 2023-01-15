/*
Copyright 2023 The Predictive Horizontal Pod Autoscaler Authors.

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

// Package scalebehavior provides functions for managing the scaling behavior of the PHPA.
// This applies HPA scaling behaviors (downscale stabilization, scale rules), PHPA scaling strategies, and min/max
// replicas.
// Much of the code for this package has been directly copied from the Kubernetes source code here:
// https://github.com/kubernetes/kubernetes/blob/3e26e104bdf9d0dc3c4046d6350b93557c67f3f4/pkg/controller/podautoscaler/horizontal.go
package scalebehavior

import (
	"fmt"
	"math"
	"sort"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

func DecideTargetReplicasByScalingStrategy(decisionType string, predictedReplicas []int32) int32 {

	// Sort in ascending order
	sort.Slice(predictedReplicas, func(i, j int) bool { return predictedReplicas[i] < predictedReplicas[j] })

	// Decide which replica count to use based on decision type
	var targetReplicas int32
	switch decisionType {
	case jamiethompsonmev1alpha1.DecisionMaximum:
		max := int32(0)
		for i, predictedReplica := range predictedReplicas {
			if i == 0 || predictedReplica > max {
				max = predictedReplica
			}
		}
		targetReplicas = max
	case jamiethompsonmev1alpha1.DecisionMinimum:
		min := int32(0)
		for i, predictedReplica := range predictedReplicas {
			if i == 0 || predictedReplica < min {
				min = predictedReplica
			}
		}
		targetReplicas = min
	case jamiethompsonmev1alpha1.DecisionMean:
		total := int32(0)
		for _, predictedReplica := range predictedReplicas {
			total += predictedReplica
		}
		if total <= 0 {
			return targetReplicas
		}
		targetReplicas = int32(math.Round(float64(total) / float64(len(predictedReplicas))))
	case jamiethompsonmev1alpha1.DecisionMedian:
		if len(predictedReplicas) <= 0 {
			return targetReplicas
		}
		halfIndex := len(predictedReplicas) / 2
		if len(predictedReplicas)%2 == 0 {
			// Even
			targetReplicas = (predictedReplicas[halfIndex-1] + predictedReplicas[halfIndex]) / 2
		} else {
			// Odd
			targetReplicas = predictedReplicas[halfIndex]
		}
	default:
		// Should not occur, panic
		panic(fmt.Errorf("unknown decision type '%s'", decisionType))
	}

	return targetReplicas
}

func DecideTargetReplicasByBehavior(
	behavior *autoscalingv2.HorizontalPodAutoscalerBehavior, currentReplicas int32, targetReplicas int32,
	minReplicas int32, maxReplicas int32,
	scaleUpReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleDownReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleUpEventHistory []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleDownEventHistory []jamiethompsonmev1alpha1.TimestampedReplicas,
	now time.Time) int32 {

	// Upscale stabilization
	upscaleMinValue := targetReplicas
	for _, timestampedReplica := range scaleUpReplicaHistory {
		if timestampedReplica.Replicas < upscaleMinValue {
			upscaleMinValue = timestampedReplica.Replicas
		}
	}

	// Downscale stabilization
	downscaleMaxValue := targetReplicas
	for _, timestampedReplica := range scaleDownReplicaHistory {
		if timestampedReplica.Replicas > downscaleMaxValue {
			downscaleMaxValue = timestampedReplica.Replicas
		}
	}

	stabilizedReplicas := currentReplicas

	if stabilizedReplicas < upscaleMinValue {
		stabilizedReplicas = upscaleMinValue
	}

	if stabilizedReplicas > downscaleMaxValue {
		stabilizedReplicas = downscaleMaxValue
	}

	targetReplicas = decideTargetReplicasByBehaviorRate(behavior, currentReplicas, stabilizedReplicas, scaleUpEventHistory,
		scaleDownEventHistory, now)

	if targetReplicas < minReplicas {
		targetReplicas = minReplicas
	}

	if targetReplicas > maxReplicas {
		targetReplicas = maxReplicas
	}

	return targetReplicas
}

// returns the longest policy period in seconds from the policies in a set of scaling rules provided
func GetLongestPolicyPeriod(scalingRules *autoscalingv2.HPAScalingRules) int32 {
	var longestPolicyPeriod int32 = 0
	if scalingRules == nil {
		return longestPolicyPeriod
	}

	for _, policy := range scalingRules.Policies {
		if policy.PeriodSeconds > longestPolicyPeriod {
			longestPolicyPeriod = policy.PeriodSeconds
		}
	}
	return longestPolicyPeriod
}

func PruneTimestampedReplicasToWindow(
	timestampedReplicas []jamiethompsonmev1alpha1.TimestampedReplicas, window int32, now time.Time) []jamiethompsonmev1alpha1.TimestampedReplicas {

	prunedTimestampedReplicas := []jamiethompsonmev1alpha1.TimestampedReplicas{}

	// Prune old evaluations
	// Cutoff is current time - stabilization window
	cutoff := now.Add(time.Duration(-window) * time.Second)

	// Add any timestamped replicas after the cut off (inside the window) to the pruned list
	for _, timestampedReplica := range timestampedReplicas {
		if timestampedReplica.Time.After(cutoff) {
			prunedTimestampedReplicas = append(prunedTimestampedReplicas, timestampedReplica)
		}
	}

	return prunedTimestampedReplicas
}

func decideTargetReplicasByBehaviorRate(behavior *autoscalingv2.HorizontalPodAutoscalerBehavior,
	currentReplicas int32, targetReplicas int32,
	scaleUpEvents []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleDownEvents []jamiethompsonmev1alpha1.TimestampedReplicas, now time.Time) int32 {

	if targetReplicas > currentReplicas {
		// Scale up
		scaleUpLimit := calculateScaleUpLimitWithinScalingRules(currentReplicas, scaleUpEvents, scaleDownEvents,
			behavior.ScaleUp, now)

		if scaleUpLimit < currentReplicas {
			// We shouldn't scale up further until the scaleUpEvents will be cleaned up
			scaleUpLimit = currentReplicas
		}

		if targetReplicas > scaleUpLimit {
			return scaleUpLimit
		}
	} else if targetReplicas < currentReplicas {
		// Scale down
		scaleDownLimit := calculateScaleDownLimitWithinScalingRules(currentReplicas, scaleUpEvents, scaleDownEvents,
			behavior.ScaleDown, now)

		if scaleDownLimit > currentReplicas {
			// We shouldn't scale down further until the scaleDownEvents will be cleaned up
			scaleDownLimit = currentReplicas
		}

		if targetReplicas < scaleDownLimit {
			return scaleDownLimit
		}
	}

	// Within the scaling limits
	return targetReplicas
}

func calculateScaleUpLimitWithinScalingRules(currentReplicas int32,
	scaleUpEvents []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleDownEvents []jamiethompsonmev1alpha1.TimestampedReplicas,
	scalingRules *autoscalingv2.HPAScalingRules, now time.Time) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscalingv2.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscalingv2.MinChangePolicySelect {
		result = math.MaxInt32
		selectPolicyFn = min // For scaling up, the lowest change ('min' policy) produces a minimum value
	} else {
		result = math.MinInt32
		selectPolicyFn = max // Use the default policy otherwise to produce a highest possible change
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicaChanges(scaleUpEvents, policy.PeriodSeconds, now)
		replicasDeletedInCurrentPeriod := getReplicaChanges(scaleDownEvents, policy.PeriodSeconds, now)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscalingv2.PodsScalingPolicy {
			proposed = periodStartReplicas + policy.Value
		} else if policy.Type == autoscalingv2.PercentScalingPolicy {
			// the proposal has to be rounded up because the proposed change might not increase the replica count causing the target to never scale up
			proposed = int32(math.Ceil(float64(periodStartReplicas) * (1 + float64(policy.Value)/100)))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

func calculateScaleDownLimitWithinScalingRules(currentReplicas int32,
	scaleUpEvents []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleDownEvents []jamiethompsonmev1alpha1.TimestampedReplicas,
	scalingRules *autoscalingv2.HPAScalingRules, now time.Time) int32 {
	var result int32
	var proposed int32
	var selectPolicyFn func(int32, int32) int32
	if *scalingRules.SelectPolicy == autoscalingv2.DisabledPolicySelect {
		return currentReplicas // Scaling is disabled
	} else if *scalingRules.SelectPolicy == autoscalingv2.MinChangePolicySelect {
		result = math.MinInt32
		selectPolicyFn = max // For scaling down, the lowest change ('min' policy) produces a maximum value
	} else {
		result = math.MaxInt32
		selectPolicyFn = min // Use the default policy otherwise to produce a highest possible change
	}
	for _, policy := range scalingRules.Policies {
		replicasAddedInCurrentPeriod := getReplicaChanges(scaleUpEvents, policy.PeriodSeconds, now)
		replicasDeletedInCurrentPeriod := getReplicaChanges(scaleDownEvents, policy.PeriodSeconds, now)
		periodStartReplicas := currentReplicas - replicasAddedInCurrentPeriod + replicasDeletedInCurrentPeriod
		if policy.Type == autoscalingv2.PodsScalingPolicy {
			proposed = periodStartReplicas - policy.Value
		} else if policy.Type == autoscalingv2.PercentScalingPolicy {
			proposed = int32(float64(periodStartReplicas) * (1 - float64(policy.Value)/100))
		}
		result = selectPolicyFn(result, proposed)
	}
	return result
}

func getReplicaChanges(scaleEvents []jamiethompsonmev1alpha1.TimestampedReplicas, periodSeconds int32, now time.Time) int32 {
	period := time.Second * time.Duration(periodSeconds)
	cutoff := now.Add(-period)
	var replicas int32
	for _, scaleEvent := range scaleEvents {
		if scaleEvent.Time.After(cutoff) {
			replicas += scaleEvent.Replicas
		}
	}
	return replicas
}

func max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}
