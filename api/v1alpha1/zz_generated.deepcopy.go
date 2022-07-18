//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/autoscaling/v2beta2"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HoltWinters) DeepCopyInto(out *HoltWinters) {
	*out = *in
	if in.Alpha != nil {
		in, out := &in.Alpha, &out.Alpha
		*out = new(float64)
		**out = **in
	}
	if in.Beta != nil {
		in, out := &in.Beta, &out.Beta
		*out = new(float64)
		**out = **in
	}
	if in.Gamma != nil {
		in, out := &in.Gamma, &out.Gamma
		*out = new(float64)
		**out = **in
	}
	if in.DampedTrend != nil {
		in, out := &in.DampedTrend, &out.DampedTrend
		*out = new(bool)
		**out = **in
	}
	if in.InitializationMethod != nil {
		in, out := &in.InitializationMethod, &out.InitializationMethod
		*out = new(string)
		**out = **in
	}
	if in.InitialLevel != nil {
		in, out := &in.InitialLevel, &out.InitialLevel
		*out = new(float64)
		**out = **in
	}
	if in.InitialTrend != nil {
		in, out := &in.InitialTrend, &out.InitialTrend
		*out = new(float64)
		**out = **in
	}
	if in.InitialSeasonal != nil {
		in, out := &in.InitialSeasonal, &out.InitialSeasonal
		*out = new(float64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HoltWinters.
func (in *HoltWinters) DeepCopy() *HoltWinters {
	if in == nil {
		return nil
	}
	out := new(HoltWinters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Linear) DeepCopyInto(out *Linear) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Linear.
func (in *Linear) DeepCopy() *Linear {
	if in == nil {
		return nil
	}
	out := new(Linear)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Model) DeepCopyInto(out *Model) {
	*out = *in
	if in.CalculationTimeout != nil {
		in, out := &in.CalculationTimeout, &out.CalculationTimeout
		*out = new(int)
		**out = **in
	}
	if in.PerSyncPeriod != nil {
		in, out := &in.PerSyncPeriod, &out.PerSyncPeriod
		*out = new(int)
		**out = **in
	}
	if in.Linear != nil {
		in, out := &in.Linear, &out.Linear
		*out = new(Linear)
		**out = **in
	}
	if in.HoltWinters != nil {
		in, out := &in.HoltWinters, &out.HoltWinters
		*out = new(HoltWinters)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Model.
func (in *Model) DeepCopy() *Model {
	if in == nil {
		return nil
	}
	out := new(Model)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ModelHistory) DeepCopyInto(out *ModelHistory) {
	*out = *in
	if in.ReplicaHistroy != nil {
		in, out := &in.ReplicaHistroy, &out.ReplicaHistroy
		*out = make([]TimestampedReplicas, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ModelHistory.
func (in *ModelHistory) DeepCopy() *ModelHistory {
	if in == nil {
		return nil
	}
	out := new(ModelHistory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PredictiveHorizontalPodAutoscaler) DeepCopyInto(out *PredictiveHorizontalPodAutoscaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PredictiveHorizontalPodAutoscaler.
func (in *PredictiveHorizontalPodAutoscaler) DeepCopy() *PredictiveHorizontalPodAutoscaler {
	if in == nil {
		return nil
	}
	out := new(PredictiveHorizontalPodAutoscaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PredictiveHorizontalPodAutoscaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PredictiveHorizontalPodAutoscalerData) DeepCopyInto(out *PredictiveHorizontalPodAutoscalerData) {
	*out = *in
	if in.ModelHistories != nil {
		in, out := &in.ModelHistories, &out.ModelHistories
		*out = make(map[string]ModelHistory, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PredictiveHorizontalPodAutoscalerData.
func (in *PredictiveHorizontalPodAutoscalerData) DeepCopy() *PredictiveHorizontalPodAutoscalerData {
	if in == nil {
		return nil
	}
	out := new(PredictiveHorizontalPodAutoscalerData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PredictiveHorizontalPodAutoscalerList) DeepCopyInto(out *PredictiveHorizontalPodAutoscalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PredictiveHorizontalPodAutoscaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PredictiveHorizontalPodAutoscalerList.
func (in *PredictiveHorizontalPodAutoscalerList) DeepCopy() *PredictiveHorizontalPodAutoscalerList {
	if in == nil {
		return nil
	}
	out := new(PredictiveHorizontalPodAutoscalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PredictiveHorizontalPodAutoscalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PredictiveHorizontalPodAutoscalerSpec) DeepCopyInto(out *PredictiveHorizontalPodAutoscalerSpec) {
	*out = *in
	out.ScaleTargetRef = in.ScaleTargetRef
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]v2beta2.MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DownscaleStabilization != nil {
		in, out := &in.DownscaleStabilization, &out.DownscaleStabilization
		*out = new(int)
		**out = **in
	}
	if in.CPUInitializationPeriod != nil {
		in, out := &in.CPUInitializationPeriod, &out.CPUInitializationPeriod
		*out = new(int)
		**out = **in
	}
	if in.InitialReadinessDelay != nil {
		in, out := &in.InitialReadinessDelay, &out.InitialReadinessDelay
		*out = new(int)
		**out = **in
	}
	if in.Tolerance != nil {
		in, out := &in.Tolerance, &out.Tolerance
		*out = new(float64)
		**out = **in
	}
	if in.SyncPeriod != nil {
		in, out := &in.SyncPeriod, &out.SyncPeriod
		*out = new(int)
		**out = **in
	}
	if in.Models != nil {
		in, out := &in.Models, &out.Models
		*out = make([]Model, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DecisionType != nil {
		in, out := &in.DecisionType, &out.DecisionType
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PredictiveHorizontalPodAutoscalerSpec.
func (in *PredictiveHorizontalPodAutoscalerSpec) DeepCopy() *PredictiveHorizontalPodAutoscalerSpec {
	if in == nil {
		return nil
	}
	out := new(PredictiveHorizontalPodAutoscalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PredictiveHorizontalPodAutoscalerStatus) DeepCopyInto(out *PredictiveHorizontalPodAutoscalerStatus) {
	*out = *in
	if in.LastScaleTime != nil {
		in, out := &in.LastScaleTime, &out.LastScaleTime
		*out = (*in).DeepCopy()
	}
	if in.ReplicaHistory != nil {
		in, out := &in.ReplicaHistory, &out.ReplicaHistory
		*out = make([]TimestampedReplicas, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CurrentMetrics != nil {
		in, out := &in.CurrentMetrics, &out.CurrentMetrics
		*out = make([]v2beta2.MetricStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PredictiveHorizontalPodAutoscalerStatus.
func (in *PredictiveHorizontalPodAutoscalerStatus) DeepCopy() *PredictiveHorizontalPodAutoscalerStatus {
	if in == nil {
		return nil
	}
	out := new(PredictiveHorizontalPodAutoscalerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TimestampedReplicas) DeepCopyInto(out *TimestampedReplicas) {
	*out = *in
	if in.Time != nil {
		in, out := &in.Time, &out.Time
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TimestampedReplicas.
func (in *TimestampedReplicas) DeepCopy() *TimestampedReplicas {
	if in == nil {
		return nil
	}
	out := new(TimestampedReplicas)
	in.DeepCopyInto(out)
	return out
}
