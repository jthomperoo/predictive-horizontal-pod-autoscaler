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

// Run (fake) provides a way to insert functionality into an algorithm Runner
type Run struct {
	RunAlgorithmWithValueReactor func(algorithmPath string, value string, timeout int) (string, error)
}

// RunAlgorithmWithValue calls the fake Runner function
func (f *Run) RunAlgorithmWithValue(algorithmPath string, value string, timeout int) (string, error) {
	return f.RunAlgorithmWithValueReactor(algorithmPath, value, timeout)
}
