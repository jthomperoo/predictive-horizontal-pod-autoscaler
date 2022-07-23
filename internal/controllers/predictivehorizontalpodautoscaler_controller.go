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

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/jthomperoo/k8shorizmetrics"
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction"
)

const (
	defaultMinReplicas             = 1
	defaultSyncPeriod              = 15 * time.Second
	defaultErrorRetryPeriod        = 10 * time.Second
	defaultDownscaleStabilization  = 300
	defaultCPUInitializationPeriod = 30
	defaultInitialReadinessDelay   = 30
	defaultTolerance               = 0.1
	defaultDecisionType            = jamiethompsonmev1alpha1.DecisionMaximum
	defaultPerSyncPeriod           = 1
)

const (
	configMapDataKey = "data"
)

// PredictiveHorizontalPodAutoscalerReconciler reconciles a PredictiveHorizontalPodAutoscaler object
type PredictiveHorizontalPodAutoscalerReconciler struct {
	client.Client
	ScaleClient scale.ScalesGetter
	Scheme      *runtime.Scheme
	Gatherer    k8shorizmetrics.Gatherer
	Evaluator   k8shorizmetrics.Evaluator
	Predicter   prediction.Predicter
}

//+kubebuilder:rbac:groups=jamiethompson.me,resources=predictivehorizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=jamiethompson.me,resources=predictivehorizontalpodautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=jamiethompson.me,resources=predictivehorizontalpodautoscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list
//+kubebuilder:rbac:groups=core,resources=replicationcontrollers/scale,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/scale;replicaset/scale;statefulset/scale,verbs=get;update;patch
//+kubebuilder:rbac:groups=metrics.k8s.io,resources=*,verbs=get;list
//+kubebuilder:rbac:groups=custom.metrics.k8s.io,resources=*,verbs=get;list
//+kubebuilder:rbac:groups=external.metrics.k8s.io,resources=*,verbs=get;list

func (r *PredictiveHorizontalPodAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get PredictiveHorizontalPodAutoscaler")
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	err = r.validate(instance)
	if err != nil {
		logger.Error(err, "invalid PredictiveHorizontalPodAutoscaler, disabling PHPA until changed to be valid")
		// We stop processing here without requeueing since the PHPA is invalid, if changes are made to the spec that
		// make it valid it will be reconciled again and the validation checked
		return reconcile.Result{}, nil
	}

	err = r.preScaleStatusCheck(ctx, instance)
	if err != nil {
		logger.Error(err, "failed pre scale status check", "scaleTargetRef", instance.Spec.ScaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	configMap, phpaData, err := r.getPHPAConfigMapAndData(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to get PHPA config map and data", "scaleTargetRef", instance.Spec.ScaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	syncPeriod := defaultSyncPeriod
	if instance.Spec.SyncPeriod != nil {
		syncPeriod = time.Duration(*instance.Spec.SyncPeriod) * time.Millisecond
	}

	now := time.Now().UTC()

	// Check the last scale of the PHPA, make sure we're not scaling too early
	lastScaleTime := instance.Status.LastScaleTime
	if lastScaleTime != nil && now.Add(-syncPeriod).Before(lastScaleTime.Time) {
		timeUntilReconcile := instance.Status.LastScaleTime.Time.Add(syncPeriod).Sub(now)
		logger.V(1).Info("Resource already scaled, queueing up reconcile for the next sync period",
			"ScaleTargetRef", instance.Spec.ScaleTargetRef,
			"syncPeriod", syncPeriod,
			"timeUntilReconcile", timeUntilReconcile.Seconds())
		return reconcile.Result{RequeueAfter: timeUntilReconcile}, nil
	}

	scale, err := r.getScaleSubresource(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to get scale subresource", "ScaleTargetRef", instance.Spec.ScaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	calculatedReplicas, err := r.calculateReplicas(instance, scale)
	if err != nil {
		logger.Error(err, "failed to calculate replicas based on metrics",
			"ScaleTargetRef", instance.Spec.ScaleTargetRef,
			"currentReplicas", scale.Spec.Replicas)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// This function doesn't return any errors, since if it fails to process a model it will skip and continue
	// processing without that model's results
	predictedReplicas, phpaData := r.processModels(ctx, instance, phpaData, now, scale.Spec.Replicas,
		calculatedReplicas)

	err = r.updateConfigMapData(ctx, configMap, phpaData)
	if err != nil {
		logger.Error(err, "failed to update PHPA configmap",
			"ScaleTargetRef", instance.Spec.ScaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	targetReplicas, downscaleStabilizationHistory, err := r.decideTargetReplicas(instance, predictedReplicas, now)
	if err != nil {
		logger.Error(err, "failed to decide targer replicas",
			"ScaleTargetRef", instance.Spec.ScaleTargetRef,
			"currentReplicas", scale.Spec.Replicas,
			"predictedReplicas", predictedReplicas,
		)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// Only scale if the current replicas is different than the target
	if scale.Spec.Replicas != targetReplicas {
		err = r.updateScaleResource(ctx, instance, scale, targetReplicas)
		if err != nil {
			logger.Error(err, "failed to update scale resource",
				"ScaleTargetRef", instance.Spec.ScaleTargetRef,
				"currentReplicas", scale.Spec.Replicas,
				"targetReplicas", targetReplicas)
			return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
		}
	}

	instance.Status.LastScaleTime = &metav1.Time{Time: now}
	instance.Status.DesiredReplicas = targetReplicas
	instance.Status.CurrentReplicas = scale.Spec.Replicas
	instance.Status.ReplicaHistory = downscaleStabilizationHistory
	err = r.Client.Status().Update(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to update status of resource",
			"ScaleTargetRef", instance.Spec.ScaleTargetRef,
			"currentReplicas", scale.Spec.Replicas,
			"targetReplicas", targetReplicas,
			"scaleTime", now)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	logger.V(0).Info("Scaled resource",
		"ScaleTargetRef", instance.Spec.ScaleTargetRef,
		"currentReplicas", scale.Spec.Replicas,
		"targetReplicas", targetReplicas)

	return reconcile.Result{RequeueAfter: syncPeriod}, nil

}

// updateScaleResource updates the replica count on the scale subresource
func (r *PredictiveHorizontalPodAutoscalerReconciler) updateScaleResource(
	ctx context.Context, instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler,
	scale *autoscalingv1.Scale, targetReplicas int32) error {
	scale.Spec.Replicas = targetReplicas

	targetGR := schema.GroupResource{
		Group:    scale.GroupVersionKind().Group,
		Resource: scale.GroupVersionKind().Kind,
	}

	_, err := r.ScaleClient.Scales(instance.Namespace).Update(ctx, targetGR, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale resource: %w", err)
	}

	return nil
}

// updateConfigMapData updates the PHPA's configmap and the data it holds
func (r *PredictiveHorizontalPodAutoscalerReconciler) updateConfigMapData(ctx context.Context, configMap *corev1.ConfigMap,
	phpaData *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerData) error {
	data, err := json.Marshal(phpaData)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	configMap.Data = map[string]string{
		configMapDataKey: string(data),
	}

	err = r.Client.Update(ctx, configMap)
	if err != nil {
		return fmt.Errorf("failed to update config map data: %w", err)
	}

	return nil
}

// decideTargetReplicas uses the list of predicted replicas (from the calculated HPA value and the model predictions)
// and returns a single value to use for scaling. This accounts for both downscale stabilization and the decisionType
// scaling strategy.
func (r *PredictiveHorizontalPodAutoscalerReconciler) decideTargetReplicas(
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler, predictedReplicas []int32,
	now time.Time) (int32, []jamiethompsonmev1alpha1.TimestampedReplicas, error) {

	minReplicas := int32(defaultMinReplicas)
	if instance.Spec.MinReplicas != nil {
		minReplicas = *instance.Spec.MinReplicas
	}

	decisionType := defaultDecisionType
	if instance.Spec.DecisionType != nil {
		decisionType = *instance.Spec.DecisionType
	}

	downscaleStabilization := defaultDownscaleStabilization
	if instance.Spec.DownscaleStabilization != nil {
		downscaleStabilization = *instance.Spec.DownscaleStabilization
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
		targetReplicas = int32(float64(int(total) / len(predictedReplicas)))
	case jamiethompsonmev1alpha1.DecisionMedian:
		halfIndex := len(predictedReplicas) / 2
		if len(predictedReplicas)%2 == 0 {
			// Even
			targetReplicas = (predictedReplicas[halfIndex-1] + predictedReplicas[halfIndex]) / 2
		} else {
			// Odd
			targetReplicas = predictedReplicas[halfIndex]
		}
	default:
		return 0, nil, fmt.Errorf("unknown decision type '%s'", decisionType)
	}

	if targetReplicas < minReplicas {
		targetReplicas = minReplicas
	}

	if targetReplicas > instance.Spec.MaxReplicas {
		targetReplicas = instance.Spec.MaxReplicas
	}

	downscaleStabilizationHistory := instance.Status.ReplicaHistory

	// Prune old evaluations
	// Cutoff is current time - stabilization window
	cutoff := &metav1.Time{Time: now.Add(time.Duration(-downscaleStabilization) * time.Second)}
	// Loop backwards over stabilization evaluations to prune old ones
	// Backwards loop to allow values to be removed mid-loop without breaking it
	for i := len(downscaleStabilizationHistory) - 1; i >= 0; i-- {
		timestampedReplica := downscaleStabilizationHistory[i]
		if timestampedReplica.Time.Before(cutoff) {
			downscaleStabilizationHistory = append(downscaleStabilizationHistory[:i], downscaleStabilizationHistory[i+1:]...)
		}
	}

	downscaleStabilizationHistory = append(downscaleStabilizationHistory, jamiethompsonmev1alpha1.TimestampedReplicas{
		Time:     &metav1.Time{Time: now},
		Replicas: targetReplicas,
	})

	for _, timestampedReplica := range downscaleStabilizationHistory {
		if timestampedReplica.Replicas > targetReplicas {
			targetReplicas = timestampedReplica.Replicas
		}
	}

	return targetReplicas, downscaleStabilizationHistory, nil
}

// processModels processes every model provided in the spec, it does not return any errors and will instead simply
// log if a model has failed to be processed, allowing the other models/the HPA calculated replicas to be used instead
func (r *PredictiveHorizontalPodAutoscalerReconciler) processModels(ctx context.Context,
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler,
	phpaData *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerData, now time.Time, currentReplicas int32,
	calculatedReplicas int32) ([]int32, *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerData) {

	logger := log.FromContext(ctx)

	scaleTargetRef := instance.Spec.ScaleTargetRef

	// Set up a slice with the calculated replicas as the first prediction
	predictedReplicas := []int32{calculatedReplicas}

	// Add the calculated replicas to a list of past replicas
	for _, model := range instance.Spec.Models {
		logger.V(2).Info("Processing model to determine replica count",
			"ScaleTargetRef", scaleTargetRef,
			"model", model.Name)

		perSyncPeriod := defaultPerSyncPeriod
		if model.PerSyncPeriod != nil {
			perSyncPeriod = *model.PerSyncPeriod
		}

		modelHistory, exists := phpaData.ModelHistories[model.Name]
		if !exists || modelHistory.Type != model.Type {
			// Create new if model doesn't exist or has a type mismatch
			modelHistory = jamiethompsonmev1alpha1.ModelHistory{
				Type:              model.Type,
				SyncPeriodsPassed: 1,
				ReplicaHistroy:    []jamiethompsonmev1alpha1.TimestampedReplicas{},
			}
		}

		shouldRunOnThisSyncPeriod := modelHistory.SyncPeriodsPassed >= perSyncPeriod

		modelHistory.ReplicaHistroy = append(modelHistory.ReplicaHistroy, jamiethompsonmev1alpha1.TimestampedReplicas{
			Time: &metav1.Time{
				Time: now,
			},
			Replicas: calculatedReplicas,
		})

		if shouldRunOnThisSyncPeriod {
			logger.V(1).Info("Using model to calculate predicted target replicas",
				"ScaleTargetRef", scaleTargetRef,
				"model", model.Name)
			replicas, err := r.Predicter.GetPrediction(&model, modelHistory.ReplicaHistroy)
			if err != nil {
				// Skip this model, errored out
				logger.Error(err, "failed to get predicted replica count",
					"ScaleTargetRef", scaleTargetRef,
					"currentReplicas", currentReplicas,
					"targetReplicas", calculatedReplicas)
				continue
			}
			predictedReplicas = append(predictedReplicas, replicas)
			modelHistory.SyncPeriodsPassed = 1
		} else {
			logger.V(1).Info("Skipping model for this sync period, should not run on this sync period",
				"ScaleTargetRef", scaleTargetRef,
				"syncPeriodsPassed", modelHistory.SyncPeriodsPassed,
				"perSyncPeriod", perSyncPeriod,
				"model", model.Name)
			modelHistory.SyncPeriodsPassed += 1
		}

		prunedHistory, err := r.Predicter.PruneHistory(&model, modelHistory.ReplicaHistroy)
		if err != nil {
			// Skip this model, errored out
			logger.Error(err, "failed to prune replica history",
				"ScaleTargetRef", scaleTargetRef)
			continue
		}

		modelHistory.ReplicaHistroy = prunedHistory
		phpaData.ModelHistories[model.Name] = modelHistory
	}

	// Delete any model data that exists without a corresponding model spec
	for modelName := range phpaData.ModelHistories {
		exists := false
		for _, model := range instance.Spec.Models {
			if modelName == model.Name {
				exists = true
				break
			}
		}

		if !exists {
			delete(phpaData.ModelHistories, modelName)
		}
	}

	return predictedReplicas, phpaData
}

// calculateReplicas does the HPA processing part of the autoscaling based on the metrics provided in the spec,
// returns the calculated value (the value the HPA would calculate based on these metrics).
func (r *PredictiveHorizontalPodAutoscalerReconciler) calculateReplicas(
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler, scale *autoscalingv1.Scale) (int32, error) {
	cpuInitializationPeriod := defaultCPUInitializationPeriod
	if instance.Spec.CPUInitializationPeriod != nil {
		cpuInitializationPeriod = *instance.Spec.CPUInitializationPeriod
	}

	initialReadinessDelay := defaultInitialReadinessDelay
	if instance.Spec.InitialReadinessDelay != nil {
		initialReadinessDelay = *instance.Spec.InitialReadinessDelay
	}

	tolerance := defaultTolerance
	if instance.Spec.Tolerance != nil {
		tolerance = *instance.Spec.Tolerance
	}

	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		return 0, fmt.Errorf("failed to parse pod selector from scale subresource selector: %w", err)
	}

	// Gather K8s metrics using the spec
	metrics, err := r.Gatherer.GatherWithOptions(instance.Spec.Metrics, scale.Namespace, selector,
		time.Duration(cpuInitializationPeriod)*time.Second, time.Duration(initialReadinessDelay)*time.Second)
	if err != nil {
		return 0, fmt.Errorf("failed to gather metrics using provided metric specs: %w", err)
	}

	// Calculate the targetReplicas using these metrics
	currentReplicas := scale.Spec.Replicas
	calculatedReplicas, err := r.Evaluator.EvaluateWithOptions(metrics, currentReplicas, tolerance)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate metrics and calculate target replica count: %w", err)
	}

	return calculatedReplicas, nil
}

// getScaleSubresource gets the scale subresource for the resource being targeted
func (r *PredictiveHorizontalPodAutoscalerReconciler) getScaleSubresource(ctx context.Context,
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler) (*autoscalingv1.Scale, error) {
	scaleTargetRef := instance.Spec.ScaleTargetRef

	// Get targeted scale subresource
	resourceGV, err := schema.ParseGroupVersion(scaleTargetRef.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group version of target resource: %w", err)
	}

	targetGR := schema.GroupResource{
		Group:    resourceGV.Group,
		Resource: scaleTargetRef.Kind,
	}

	scale, err := r.ScaleClient.Scales(instance.Namespace).Get(ctx, targetGR, scaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("ailed to get the scale subresource of the target resource: %w", err)
	}

	return scale, nil
}

// getPHPAConfigMapAndData returns the config map and parsed data for the PHPA
func (r *PredictiveHorizontalPodAutoscalerReconciler) getPHPAConfigMapAndData(ctx context.Context,
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler) (*corev1.ConfigMap, *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerData, error) {

	logger := log.FromContext(ctx)

	scaleTargetRef := instance.Spec.ScaleTargetRef

	// Check if configmap exists, if not create a blank one
	phpaData := &jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscalerData{
		ModelHistories: map[string]jamiethompsonmev1alpha1.ModelHistory{},
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("predictive-horizontal-pod-autoscaler-%s-data", instance.Name),
			Namespace: instance.Namespace,
		},
	}
	configMap.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: instance.APIVersion,
		Kind:       instance.Kind,
		Name:       instance.Name,
		UID:        instance.UID,
	}})

	err := r.Client.Get(context.Background(), types.NamespacedName{Name: configMap.GetName(), Namespace: configMap.GetNamespace()}, configMap)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, nil, fmt.Errorf("failed to get PHPA configmap: %w", err)
		}

		logger.V(1).Info("No configmap found for PHPA, creating a new one",
			"ScaleTargetRef", scaleTargetRef)

		data, err := json.Marshal(phpaData)
		if err != nil {
			// Should not occur, panic
			panic(err)
		}

		configMap.Data = map[string]string{
			configMapDataKey: string(data),
		}

		err = r.Client.Create(ctx, configMap)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create PHPA configmap: %w", err)
		}
	}

	err = json.Unmarshal([]byte(configMap.Data[configMapDataKey]), phpaData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse PHPA data: %w", err)
	}

	return configMap, phpaData, nil
}

// preScaleStatusCheck makes sure that the PHPAs status fields are correct before scaling, e.g. the reference field
// is set
func (r *PredictiveHorizontalPodAutoscalerReconciler) preScaleStatusCheck(ctx context.Context,
	instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler) error {

	scaleTargetRef := instance.Spec.ScaleTargetRef

	reference := fmt.Sprintf("%s/%s", scaleTargetRef.Kind, scaleTargetRef.Name)
	if instance.Status.Reference != reference {
		instance.Status.Reference = reference
		err := r.Client.Status().Update(ctx, instance)
		if err != nil {
			return fmt.Errorf("failed to update status of resource: %w", err)
		}
	}

	return nil
}

// validate performs validation on the PHPA, will return an error if the PHPA is not valid
func (r *PredictiveHorizontalPodAutoscalerReconciler) validate(instance *jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler) error {
	spec := instance.Spec

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

	for _, model := range spec.Models {
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

// SetupWithManager sets up the controller with the Manager.
func (r *PredictiveHorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
