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
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	scaleTargetRef := instance.Spec.ScaleTargetRef

	reference := fmt.Sprintf("%s/%s", scaleTargetRef.Kind, scaleTargetRef.Name)
	if instance.Status.Reference != reference {
		instance.Status.Reference = reference
		err = r.Client.Status().Update(ctx, instance)
		if err != nil {
			logger.Error(err, "failed to update status of resource",
				"ScaleTargetRef", scaleTargetRef,
				"reference", reference)
			return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
		}
	}

	// TODO: add PHPA validation

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

	err = r.Client.Get(context.Background(), types.NamespacedName{Name: configMap.GetName(), Namespace: configMap.GetNamespace()}, configMap)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "failed to get PHPA configmap",
				"ScaleTargetRef", scaleTargetRef)
			// If it's an error other than not found, requeue with the error
			return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
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
			logger.Error(err, "failed to create PHPA configmap",
				"ScaleTargetRef", scaleTargetRef)
			return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
		}
	}

	err = json.Unmarshal([]byte(configMap.Data[configMapDataKey]), phpaData)
	if err != nil {
		logger.Error(err, "failed to parse PHPA data",
			"ScaleTargetRef", scaleTargetRef,
			"data", configMap.Data[configMapDataKey])
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
		// We've already scaled before the last sync period, so lets just wait until that time before reconciling again
		timeUntilReconcile := lastScaleTime.Time.Add(syncPeriod).Sub(now)
		logger.V(1).Info("Resource already scaled, queueing up reconcile for the next sync period",
			"ScaleTargetRef", scaleTargetRef,
			"syncPeriod", syncPeriod,
			"timeUntilReconcile", timeUntilReconcile.Seconds())
		return reconcile.Result{RequeueAfter: timeUntilReconcile}, nil
	}

	minReplicas := int32(defaultMinReplicas)
	if instance.Spec.MinReplicas != nil {
		minReplicas = *instance.Spec.MinReplicas
	}

	downscaleStabilization := defaultDownscaleStabilization
	if instance.Spec.DownscaleStabilization != nil {
		downscaleStabilization = *instance.Spec.DownscaleStabilization
	}

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

	decisionType := defaultDecisionType
	if instance.Spec.DecisionType != nil {
		decisionType = *instance.Spec.DecisionType
	}

	// Get targeted scale subresource
	resourceGV, err := schema.ParseGroupVersion(scaleTargetRef.APIVersion)
	if err != nil {
		logger.Error(err, "failed to parse group version of target resource",
			"ScaleTargetRef", scaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	targetGR := schema.GroupResource{
		Group:    resourceGV.Group,
		Resource: scaleTargetRef.Kind,
	}

	scale, err := r.ScaleClient.Scales(instance.Namespace).Get(ctx, targetGR, scaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get the scale subresource of the target resource",
			"ScaleTargetRef", scaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		logger.Error(err, "failed to parse pod selector from scale subresource selector",
			"ScaleTargetRef", scaleTargetRef,
			"selector", scale.Status.Selector)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// Gather K8s metrics using the spec
	metrics, err := r.Gatherer.GatherWithOptions(instance.Spec.Metrics, scale.Namespace, selector,
		time.Duration(cpuInitializationPeriod)*time.Second, time.Duration(initialReadinessDelay)*time.Second)
	if err != nil {
		logger.Error(err, "failed to gather metrics using provided metric specs",
			"ScaleTargetRef", scaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	// Calculate the targetReplicas using these metrics
	currentReplicas := scale.Spec.Replicas
	calculatedReplicas, err := r.Evaluator.EvaluateWithOptions(metrics, currentReplicas, tolerance)
	if err != nil {
		logger.Error(err, "failed to evaluate metrics and calculate target replica count",
			"ScaleTargetRef", scaleTargetRef,
			"currentReplicas", currentReplicas)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

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

		modelHistory.ReplicaHistroy, err = r.Predicter.PruneHistory(&model, modelHistory.ReplicaHistroy)
		if err != nil {
			// Skip this model, errored out
			logger.Error(err, "failed to prune replica history",
				"ScaleTargetRef", scaleTargetRef)
			continue
		}

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

	// Config map not found, create a new one
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
		logger.Error(err, "failed to update config map data",
			"ScaleTargetRef", scaleTargetRef)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
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
		// Should not occur, panic with an error
		panic(fmt.Errorf("unknown decision type '%s'", decisionType))
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

	if currentReplicas != targetReplicas {
		scale.Spec.Replicas = targetReplicas
		_, err = r.ScaleClient.Scales(instance.Namespace).Update(ctx, targetGR, scale, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(err, "failed to scale resource",
				"ScaleTargetRef", scaleTargetRef,
				"currentReplicas", currentReplicas,
				"targetReplicas", targetReplicas)
			return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
		}
	}

	instance.Status.LastScaleTime = &metav1.Time{Time: now}
	instance.Status.DesiredReplicas = targetReplicas
	instance.Status.CurrentReplicas = currentReplicas
	instance.Status.ReplicaHistory = downscaleStabilizationHistory
	err = r.Client.Status().Update(ctx, instance)
	if err != nil {
		logger.Error(err, "failed to update status of resource",
			"ScaleTargetRef", scaleTargetRef,
			"currentReplicas", currentReplicas,
			"targetReplicas", targetReplicas,
			"scaleTime", now)
		return reconcile.Result{RequeueAfter: defaultErrorRetryPeriod}, err
	}

	logger.V(0).Info("Scaled resource",
		"ScaleTargetRef", scaleTargetRef,
		"currentReplicas", currentReplicas,
		"targetReplicas", targetReplicas)

	return reconcile.Result{RequeueAfter: syncPeriod}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *PredictiveHorizontalPodAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jamiethompsonmev1alpha1.PredictiveHorizontalPodAutoscaler{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
