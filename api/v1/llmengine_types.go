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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LLMEngineType defines the type of LLM engine.
// +kubebuilder:validation:Enum=ollama
type LLMEngineType string

const (
	// LLMEngineTypeOllama represents the Ollama engine.
	LLMEngineTypeOllama LLMEngineType = "ollama"
)

// LLMEngineSpec defines the desired state of LLMEngine.
type LLMEngineSpec struct {
	// Type specifies the type of LLM engine (e.g., ollama).
	// +kubebuilder:validation:Required
	Type LLMEngineType `json:"type"`

	// UseGPU specifies whether to use GPU for the engine.
	// +kubebuilder:default:=false
	// +optional
	UseGPU bool `json:"useGPU,omitempty"`

	// HTTPPort specifies the open HTTP port for the engine.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=8080
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// Image specifies the Container image to use for the engine.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Args specifies the arguments to pass to the Container.
	// +optional
	Args []string `json:"args,omitempty"`

	// ModelPath specifies the path where the models are located.
	// This can be a mounted volume from a PVC.
	// +kubebuilder:validation:Required
	ModelPath string `json:"modelPath"`
}

// LLMEngineStatus defines the observed state of LLMEngine.
type LLMEngineStatus struct {
	// Conditions represent the latest available observations of the LLMEngine's state.
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Ready indicates whether the LLMEngine is ready to serve requests.
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LLMEngine is the Schema for the llmengines API.
type LLMEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMEngineSpec   `json:"spec,omitempty"`
	Status LLMEngineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LLMEngineList contains a list of LLMEngine.
type LLMEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMEngine{}, &LLMEngineList{})
}
