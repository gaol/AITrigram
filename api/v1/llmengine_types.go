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

// ModelStorage defines the storage mount for the LLM model files.
type ModelStorage struct {
	// the path which will be mounted inside of the container
	// +kubebuilder:validation:Required
	Path string `json:"path"`
	// the VolumeSource from which the storage comes from
	corev1.VolumeSource `json:",inline"`
}

type CacheStorage struct {
	// the mount path in the container for cache
	// +kubebuilder:validation:Required
	Path string `json:"path"`
	// the mount path in the container
	*corev1.EmptyDirVolumeSource `json:",inline"`
}

type LLMEngineStorage struct {
	// This is the storage configuration for the k-v cache set up
	// +optional
	CacheStorage *CacheStorage `json:"cache,omitempty"`
	// This presents where the LLM files are loaded from
	// +optional
	ModelsStorage *ModelStorage `json:"models,omitempty"`
}

// LLMEngineType defines the type of LLM engine.
// +kubebuilder:validation:Enum=ollama;vllm
type LLMEngineType string

const (
	LLMEngineTypeOllama LLMEngineType = "ollama"
	LLMEngineTypeVLLM   LLMEngineType = "vllm"
)

// LLMEngineSpec defines the desired state of LLMEngine.
type LLMEngineSpec struct {
	// Type specifies the type of LLM engine (e.g., ollama, vllm).
	// +kubebuilder:validation:Required
	EngineType LLMEngineType `json:"engineType"`

	// Image specifies the container image to use for the engine.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Port specifies the open HTTP port for the engine inside of the container
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Port int32 `json:"port,omitempty"`

	// ServicePort specifies the port for the ClusterIP Service managed by this engine.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// ModelDeploymentTemplate provides default values for LLMModel CRs.
	// +optional
	ModelDeploymentTemplate *ModelDeploymentTemplate `json:"modelDeploymentTemplate,omitempty"`
}

type ModelDeploymentTemplate struct {

	// Default arguments to start the engine container.
	// +optional
	Args []string `json:"args,omitempty"`

	// Environment variables for the model container.
	// +optional
	Envs *[]corev1.EnvVar `json:"env,omitempty"`

	// Storage specifies where the models are found and loaded
	// +optional
	Storage *LLMEngineStorage `json:"storage,omitempty"`

	// DownloadImage for model preparation
	// +optional
	DownloadImage string `json:"downloadImage,omitempty"`

	// DownloadScripts for model preparation
	// +optional
	DownloadScripts string `json:"downloadScripts,omitempty"`
}

// LLMEngineStatus defines the observed state of LLMEngine.
type LLMEngineStatus struct {
	// Conditions represent the latest available observations of the LLMEngine's state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Ready      bool               `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type LLMEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMEngineSpec   `json:"spec,omitempty"`
	Status LLMEngineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type LLMEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMEngine{}, &LLMEngineList{})
}
