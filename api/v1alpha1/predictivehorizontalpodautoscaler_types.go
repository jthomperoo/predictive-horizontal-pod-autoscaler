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

package v1alpha1

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DecisionMaximum means use the highest predicted value from the models
	DecisionMaximum = "maximum"
	// DecisionMinimum means use the lowest predicted value from the models
	DecisionMinimum = "minimum"
	// DecisionMean means use the mean average of predicted values
	DecisionMean = "mean"
	// DecisionMedian means use the median average of predicted values
	DecisionMedian = "median"
)

const (
	TypeHoltWinters = "HoltWinters"
	TypeLinear      = "Linear"
)

const (
	HookTypeHTTP = "http"
)

// HookDefinition describes a hook for passing data/triggering logic, such as through a shell command
type HookDefinition struct {
	// +kubebuilder:validation:Enum=http
	Type string `json:"type"`
	// +kubebuilder:validation:Minimum=1
	Timeout int `json:"timeout"`
	// +optional
	HTTP *HTTPHook `json:"http"`
}

// HTTPHook describes configuration options for an HTTP request hook
type HTTPHook struct {
	// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;CONNECT;OPTIONS;TRACE;PATCH
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers,omitempty"`
	SuccessCodes []int             `json:"successCodes"`
	// +kubebuilder:validation:Enum=query;body
	ParameterMode string `json:"parameterMode"`
}

// Linear represents a linear regression prediction model configuration
type Linear struct {
	// historySize is how many timestamped replica counts should be stored for this linear regression, with older
	// timestamped replica counts being removed from the data as new ones are added. For example a value of 6 means
	// there will only be a maxmimu of 6 stored timestamped replica counts for this model.
	// +kubebuilder:validation:Minimum=1
	HistorySize int `json:"historySize"`
	// lookAhead is how far in the future should the linear regression predict in seconds. For example a value of 10
	// will predict 10 seconds into the future
	// +kubebuilder:validation:Minimum=1
	LookAhead int `json:"lookAhead"`
}

// HoltWinters represents a holt-winters exponential smoothing prediction model configuration
type HoltWinters struct {
	// +kubebuilder:validation:Minimum=0
	// +optional
	Alpha *float64 `json:"alpha"`

	// +kubebuilder:validation:Minimum=0
	// +optional
	Beta *float64 `json:"beta"`

	// +kubebuilder:validation:Minimum=0
	// +optional
	Gamma *float64 `json:"gamma"`

	// +kubebuilder:validation:Enum=add;additive;mul;multiplicative
	Trend string `json:"trend"`

	// +kubebuilder:validation:Enum=add;additive;mul;multiplicative
	Seasonal string `json:"seasonal"`

	// +kubebuilder:validation:Minimum=1
	SeasonalPeriods int `json:"seasonalPeriods"`

	// +kubebuilder:validation:Minimum=1
	StoredSeasons int `json:"storedSeasons"`

	// +optional
	DampedTrend *bool `json:"dampedTrend"`

	// +optional
	// +kubebuilder:validation:Enum=estimated;heuristic;known;legacy-heuristic
	InitializationMethod *string `json:"initializationMethod"`

	// +optional
	InitialLevel *float64 `json:"initialLevel"`

	// +optional
	InitialTrend *float64 `json:"initialTrend"`

	// +optional
	InitialSeasonal *float64 `json:"initialSeasonal"`

	// +optional
	RuntimeTuningFetchHook *HookDefinition `json:"runtimeTuningFetchHook"`
}

// Model represents a prediction model to use, e.g. a linear regression
type Model struct {
	// type is the type of the model, for example 'Linear'. To see a full list of supported model types visit
	// https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/user-guide/models/.
	// +kubebuilder:validation:Enum=Linear;HoltWinters
	Type string `json:"type"`

	// name is the name of the model, this can be any arbitrary name and is just used to distinguish between models if
	// you have multiple and to keep track of model data if you modify your model parameters.
	Name string `json:"name"`

	// calculationTimeout is how long the PHPA should allow for the model to calculate a value in milliseconds, if it
	// takes longer than this timeout it should skip processing the model.
	// Default varies based on model type:
	// Linear is 30000 milliseconds (30 seconds)
	// +kubebuilder:validation:Minimum=1
	// +optional
	CalculationTimeout *int `json:"calculationTimeout"`

	// perSyncPeriod is how frequently this model will run, with the syncPeriod as a base unit. This allows for you to
	// have multiple models which run at different time intervals, or only run the model every x number of sync periods
	// if the model is computation intensive.
	// For sync periods that the model is not run on, it will still add the calculated replica values to the model data
	// history and then prune that history if needs.
	// Default value is 1 (run every sync period)
	// +kubebuilder:validation:Minimum=1
	// +optional
	PerSyncPeriod *int `json:"perSyncPeriod"`

	// linear is the configuration to use for the linear regression model, it will only be used if the type is set to
	// 'Linear'.
	// +optional
	Linear *Linear `json:"linear"`

	// holtWinters is the configuration to use for the holt winters model, it will only be used if the type is set to
	// 'HoltWinters'
	// +optional
	HoltWinters *HoltWinters `json:"holtWinters"`
}

// TimestampedReplicas is a replica count paired with the time that the replica count was created at.
type TimestampedReplicas struct {
	// time is the time that the replica count was created at.
	Time *metav1.Time `json:"time"`
	// replicas is the replica count at the time.
	Replicas int32 `json:"replicas"`
}

// PredictiveHorizontalPodAutoscalerData is the data storage format for the PHPA, this is stored in a ConfigMap
type PredictiveHorizontalPodAutoscalerData struct {
	// modelHistories is a mapping of model names to model histories. This allows looking up a model's model history,
	// while allowing all of the model histories for a single PHPA to be stored in a single place.
	ModelHistories map[string]ModelHistory `json:"modelHistories"`
}

// ModelHistory is the data stored for a single model's history, with all of the replica data the model is fed to
// calculate replicas.
type ModelHistory struct {
	// type is the type of the model the history is for, useful to check for model mismatches with the data.
	Type string `json:"type"`
	// syncPeriodsPassed is the number of sync periods that have passed since the last time this model was used, used
	// when determining if a model should run based on the perSyncPeriod
	SyncPeriodsPassed int `json:"syncPeriodsPassed"`
	// replicaHistory is a list of timestamped replicas, this data is fed into the model to calculate a predicted value.
	ReplicaHistroy []TimestampedReplicas `json:"replicaHistory"`
}

// PredictiveHorizontalPodAutoscalerSpec defines the desired state of PredictiveHorizontalPodAutoscaler
type PredictiveHorizontalPodAutoscalerSpec struct {
	// scaleTargetRef points to the target resource to scale, and is used to the pods for which metrics
	// should be collected, as well as to actually change the replica count.
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference `json:"scaleTargetRef"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if at least one Object or
	// External metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +kubebuilder:validation:Minimum=0
	// +optional
	MinReplicas *int32 `json:"minReplicas"`

	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less than minReplicas.
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas"`

	// metrics contains the specifications for which to use to calculate the desired replica count (the maximum replica
	// count across all metrics will be used).  The desired replica count is calculated multiplying the ratio between
	// the target value and the current value by the current number of pods.  Ergo, metrics used must decrease as the
	// pod count is increased, and vice-versa.  See the individual metric source types for more information about how
	// each type of metric must respond. If not set, the default metric will be set to 80% average CPU utilization.
	// +listType=atomic
	// +optional
	Metrics []autoscalingv2.MetricSpec `json:"metrics"`

	// downscaleStabilization defines in seconds the length of the downscale stabilization window; based on the
	// Horizontal Pod Autoscaler downscale stabilization. Downscale stabilization works by recording all evaluations
	// over the window specified and picking out the maximum target replicas from these evaluations. This results in a
	// more smoothed downscaling and a cooldown, which can reduce the effect of thrashing.
	// Default value 300 seconds (5 minutes).
	// +kubebuilder:validation:Minimum=0
	// +optional
	DownscaleStabilization *int `json:"downscaleStabilization"`

	// cpuInitializationPeriod is equivalent to --horizontal-pod-autoscaler-cpu-initialization-period; the period after
	// pod start when CPU samples might be skipped.
	// Default value 300 seconds (5 minutes).
	// +kubebuilder:validation:Minimum=0
	// +optional
	CPUInitializationPeriod *int `json:"cpuInitializationPeriod"`

	// initialReadinessDelay is equivalent to --horizontal-pod-autoscaler-initial-readiness-delay; the period after pod
	// start during which readiness changes will be treated as initial readiness.
	// Default value 30 seconds.
	// +kubebuilder:validation:Minimum=0
	// +optional
	InitialReadinessDelay *int `json:"initialReadinessDelay"`

	// tolerance is equivalent to --horizontal-pod-autoscaler-tolerance; the minimum change (from 1.0) in the
	// desired-to-actual metrics ratio for the predictive horizontal pod autoscaler to consider scaling.
	// Default value 0.1.
	// +kubebuilder:validation:Minimum=0
	// +optional
	Tolerance *float64 `json:"tolerance"`

	// syncPeriod is equivalent to --horizontal-pod-autoscaler-sync-period; the frequency with which the PHPA
	// calculates replica counts and scales in milliseconds.
	// Default value 15000 milliseconds (15 seconds).
	// +kubebuilder:validation:Minimum=1
	// +optional
	SyncPeriod *int `json:"syncPeriod"`

	// models is the list of models to apply to the calculated replica count to calculate predicted replica values.
	// +kubebuilder:validation:Required
	Models []Model `json:"models"`

	// decisionType is the strategy to use when picking which replica count to use if you have multiple models, or even
	// just choosing between the calculculated replicas and the predicted replicas of a single model. For details on
	// which decisionTypes are available visit
	// https://predictive-horizontal-pod-autoscaler.readthedocs.io/en/latest/reference/configuration/#decisiontype
	// Default strategy is 'maximum'
	// +kubebuilder:validation:Enum=maximum;minimum;mean;median
	// +optional
	DecisionType *string `json:"decisionType"`
}

// PredictiveHorizontalPodAutoscalerStatus defines the observed state of PredictiveHorizontalPodAutoscaler
type PredictiveHorizontalPodAutoscalerStatus struct {
	// lastScaleTime is the last time the PredictiveHorizontalPodAutoscaler scaled the number of pods,
	// used by the autoscaler to control how often the number of pods is changed.
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	// replicaHistory is a timestamped history of all the calculated replica values, used for calculating downscale
	// stabilization.
	// +optional
	ReplicaHistory []TimestampedReplicas `json:"replicaHistory"`

	// reference is the resource being referenced and targeted for scaling.
	Reference string `json:"reference"`

	// currentReplicas is current number of replicas of pods managed by this autoscaler,
	// as last seen by the autoscaler.
	// +optional
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// desiredReplicas is the desired number of replicas of pods managed by this autoscaler,
	// as last calculated by the autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas"`

	// currentMetrics is the last read state of the metrics used by this autoscaler.
	// +listType=atomic
	// +optional
	CurrentMetrics []autoscalingv2.MetricStatus `json:"currentMetrics"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=phpa
// +kubebuilder:printcolumn:name="Reference",type="string",JSONPath=`.status.reference`,description="The identifier for the resource being scaled in the format <api-version>/<api-kind/<name>"
// +kubebuilder:printcolumn:name="Min Pods",type="integer",JSONPath=`.spec.minReplicas`,description="The minimum number of replicas of pods that the resource being managed by the autoscaler can have"
// +kubebuilder:printcolumn:name="Max Pods",type="integer",JSONPath=`.spec.maxReplicas`,description="The maximum number of replicas of pods that the resource being managed by the autoscaler can have"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=`.status.desiredReplicas`,description="The desired number of replicas of pods managed by this autoscaler as last calculated by the autoscaler"
// +kubebuilder:printcolumn:name="Last Scale Time",type="date",JSONPath=`.status.lastScaleTime`,description="The last time the PredictiveHorizontalPodAutoscaler scaled the number of pods"
// PredictiveHorizontalPodAutoscaler is the Schema for the predictivehorizontalpodautoscalers API
type PredictiveHorizontalPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PredictiveHorizontalPodAutoscalerSpec   `json:"spec,omitempty"`
	Status PredictiveHorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PredictiveHorizontalPodAutoscalerList contains a list of PredictiveHorizontalPodAutoscaler
type PredictiveHorizontalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PredictiveHorizontalPodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PredictiveHorizontalPodAutoscaler{}, &PredictiveHorizontalPodAutoscalerList{})
}
