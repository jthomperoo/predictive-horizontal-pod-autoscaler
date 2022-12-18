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

package scalebehavior

import (
	"fmt"
	"math"
	"sort"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

const (
	defaultDecisionType = jamiethompsonmev1alpha1.DecisionMaximum

	defaultMinReplicas = 1

	scaleUpLimitFactor  = 2.0
	scaleUpLimitMinimum = 4.0
)

func DecideTargetReplicasByScalingStrategy(
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler, predictedReplicas []int32) int32 {

	decisionType := defaultDecisionType
	if instance.Spec.DecisionType != nil {
		decisionType = *instance.Spec.DecisionType
	}

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
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler, currentReplicas int32, targetReplicas int32,
	scaleDownReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas,
	scaleUpReplicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) int32 {

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

	if instance.Spec.Behavior == nil {
		return decideTargetReplicasByDefaultBehaviorRate(instance, currentReplicas, stabilizedReplicas)
	}

	return decideTargetReplicasByBehaviorRate(instance, currentReplicas, stabilizedReplicas)
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

func decideTargetReplicasByDefaultBehaviorRate(instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler,
	currentReplicas int32, targetReplicas int32) int32 {

	if targetReplicas > currentReplicas {
		// Scale up
		maxReplicas := instance.Spec.MaxReplicas

		scaleUpLimit := calculateDefaultScaleUpLimit(currentReplicas)

		if maxReplicas > scaleUpLimit {
			maxReplicas = scaleUpLimit
		}

		if targetReplicas > maxReplicas {
			targetReplicas = maxReplicas
		}
	} else if targetReplicas < currentReplicas {
		// Scale down
		minReplicas := int32(defaultMinReplicas)
		if instance.Spec.MinReplicas != nil {
			minReplicas = *instance.Spec.MinReplicas
		}

		if targetReplicas < minReplicas {
			targetReplicas = minReplicas
		}
	}

	return targetReplicas
}

func decideTargetReplicasByBehaviorRate(instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler,
	currentReplicas int32, targetReplicas int32) int32 {
	// TODO: Implement this, see
	// https://github.com/kubernetes/kubernetes/blob/3e26e104bdf9d0dc3c4046d6350b93557c67f3f4/pkg/controller/podautoscaler/horizontal.go#L1049
	return targetReplicas
}

func calculateDefaultScaleUpLimit(currentReplicas int32) int32 {
	return int32(math.Max(scaleUpLimitFactor*float64(currentReplicas), scaleUpLimitMinimum))
}
