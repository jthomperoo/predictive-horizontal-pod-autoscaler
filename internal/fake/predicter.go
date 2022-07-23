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

package fake

import (
	jamiethompsonmev1alpha1 "github.com/jthomperoo/predictive-horizontal-pod-autoscaler/api/v1alpha1"
)

// Predicter (fake) provides a way to insert functionality into a Predicter
type Predicter struct {
	GetPredictionReactor func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error)
	PruneHistoryReactor  func(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error)
	GetTypeReactor       func() string
}

// GetIDsToRemove calls the fake Predicter function
func (f *Predicter) PruneHistory(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) ([]jamiethompsonmev1alpha1.TimestampedReplicas, error) {
	return f.PruneHistoryReactor(model, replicaHistory)
}

// GetPrediction calls the fake Predicter function
func (f *Predicter) GetPrediction(model *jamiethompsonmev1alpha1.Model, replicaHistory []jamiethompsonmev1alpha1.TimestampedReplicas) (int32, error) {
	return f.GetPredictionReactor(model, replicaHistory)
}

// GetType calls the fake Predicter function
func (f *Predicter) GetType() string {
	return f.GetTypeReactor()
}
