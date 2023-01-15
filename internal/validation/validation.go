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

package validation

import (
	"errors"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"

	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

// Validate performs validation on the PHPA, will return an error if the PHPA is not valid
func Validate(instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler) error {
	spec := instance.Spec

	err := validateMinMax(spec)
	if err != nil {
		return err
	}

	err = validateModels(spec.Models)
	if err != nil {
		return err
	}

	return nil
}

func validateMinMax(spec jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerSpec) error {
	if spec.MinReplicas != nil && spec.MaxReplicas < *spec.MinReplicas {
		return fmt.Errorf("spec.maxReplicas (%d) cannot be less than spec.minReplicas (%d)",
			spec.MaxReplicas, *spec.MinReplicas)
	}

	if spec.MinReplicas != nil && *spec.MinReplicas == 0 {
		// We need to check that if they set min replicas to zero they have at least 1 object or external metric
		// configured
		valid := false
		for _, metric := range spec.Metrics {
			if metric.Type == autoscalingv2.ObjectMetricSourceType || metric.Type == autoscalingv2.ExternalMetricSourceType {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("spec.minReplicas can only be 0 if you have at least 1 object or external metric configured")
		}
	}
	return nil
}

func validateModels(models []jamiethompsonmev1alpha1.Model) error {
	for _, model := range models {
		if model.Type == jamiethompsonmev1alpha1.TypeHoltWinters {
			hw := model.HoltWinters
			if hw == nil {
				return fmt.Errorf("invalid model '%s', type is '%s' but no Holt Winters configuration provided",
					model.Name, model.Type)
			}

			if hw.RuntimeTuningFetchHook != nil {
				hook := hw.RuntimeTuningFetchHook
				if hook.Type == jamiethompsonmev1alpha1.HookTypeHTTP && hook.HTTP == nil {
					return fmt.Errorf("invalid model '%s', runtimeTuningFetchHook is type '%s' but no HTTP hook configuration provided",
						model.Name, hook.Type)
				}
			}
		}

		if model.Type == jamiethompsonmev1alpha1.TypeLinear && model.Linear == nil {
			return fmt.Errorf("invalid model '%s', type is '%s' but no Linear Regression configuration provided",
				model.Name, model.Type)
		}
	}
	return nil
}
