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

// LLMEngineType defines the type of LLM engine.
// +kubebuilder:validation:Enum=ollama
type LLMEngineType string

const (
	// LLMEngineTypeOllama represents the Ollama engine.
	LLMEngineTypeOllama LLMEngineType = "ollama"
)

// This defines the storage mount for the pods of LLMEngine
// The VolumeSource comes from the core api, which can be any type of VolumeSource.
type StroageMount struct {

	// the path which will be mounted inside of the container
	// +kubebuilder:validation:Required
	Path string `json:"path"`

	// the VolumeSource from which the storage comes from
	corev1.VolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

type LLMEngineStorage struct {

	// This is the stroage configuration for the k-v cache set up
	// The purpose of using k-v cache in LLM inference is to improve the performance
	// LLM inference works without this, but having it may increase the throughput
	// It is possible to have bigger data than the LLM itself.
	// +optional
	CacheStorage *StroageMount `json:"cacheDir,omitempty"`

	// This presents where the LLM files are loaded from
	// It can be mounted as ReadOnlyMany as normally the LLM files won't be updated after downloaded.
	// +optional
	ModelsStorage *StroageMount `json:"modelsDir,omitempty"`
}

// LLMEngineSpec defines the desired state of LLMEngine.
type LLMEngineSpec struct {
	// Type specifies the type of LLM engine (e.g., ollama).
	// +kubebuilder:validation:Required
	EngineType *LLMEngineType `json:"engineType"`

	// HTTPPort specifies the open HTTP port for the engine.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	HTTPPort *int32 `json:"httpPort,omitempty"`

	// ServicePort specifies the port for the ClusterIP Service associated to this LLMEngine.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=8080
	// +optional
	ServicePort *int32 `json:"servicePort,omitempty"`

	// Image specifies the Container image to use for the engine.
	// It is optional because the controller will use the builtin one based on the supported EngineType.
	// Users are open to specify their own image, but that is not recommended unless they know what to do.
	// +optional
	Image *string `json:"image"`

	// Args specifies the arguments to pass to the Container.
	// +optional
	Args *[]string `json:"args,omitempty"`

	// Stroage specifies where the models are found and loaded
	// and also the cache volume to speed up the LLM inference by some frameworks like vLLM
	// +optional
	Stroage *LLMEngineStorage `json:"storage"`

	// Replicas specifies how many replicas the LLMEngine will be deployed.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Envs specifies the environment variables to pass to the Container.
	// +optional
	Envs *[]corev1.EnvVar `json:"env,omitempty"`
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
