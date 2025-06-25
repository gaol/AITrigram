/*
Copyright 2025 Lin Gao.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LLMModelSpec defines the desired state of LLMModel.
type LLMModelSpec struct {
	// Name specifies the LLM model name.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// It is the name inside of the engine itself, it is the Name if not defined.
	// +optional
	NameInEngine string `json:"nameInEngine,omitempty"`

	// EngineRef refers to the LLMEngine where this LLMModel will be deployed into
	// +kubebuilder:validation:Required
	EngineRef string `json:"engineRef"`

	// Number of replicas for this model.
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`

	// Resource requirements for this model (overrides engine defaults).
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// ModelDeployment sets up how the LLMModel will be deployed in a Deployment
	// the values here will override the default values from the LLMEngine.
	// This is useful when you want to deploy a model with different settings than the engine.
	// +optional
	ModelDeployment *ModelDeploymentTemplate `json:"modelDeployment,omitempty"`
}

// LLMModelStatus defines the observed state of LLMModel.
type LLMModelStatus struct {
	// Conditions represent the latest available observations of the LLMModel's state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Ready      bool               `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LLMModel is the Schema for the llmmodels API.
type LLMModel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMModelSpec   `json:"spec,omitempty"`
	Status LLMModelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LLMModelList contains a list of LLMModel.
type LLMModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMModel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMModel{}, &LLMModelList{})
}
