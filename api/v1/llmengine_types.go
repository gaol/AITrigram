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

// This defines the storage mount for the LLM model files.
// Typically it can be read only mode
// The VolumeSource comes from the core api, which can be any type of VolumeSource.
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
	// The purpose of using k-v cache in LLM inference is to improve the performance
	// LLM inference works without this, but having it may increase the throughput
	// It is possible to have bigger data than the LLM itself.
	// Multiple pods within a deployment should not use the same volume mount to avoid race conditions on the cache
	// Currently it only supports EmptyDirVolumeSource so that each pod instance creates it's own cache
	// +optional
	CacheStorage *CacheStorage `json:"cache,omitempty"`

	// This presents where the LLM files are loaded from
	// It can be mounted as ReadOnlyMany as normally the LLM files won't be updated after downloaded.
	// +optional
	ModelsStorage *ModelStorage `json:"models,omitempty"`
}

// ModelSpec defines a LLM model serving
type ModelSpec struct {

	// Name specifies the LLM model name in the report, it does not need to match the name inside of the engine
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// It is the name inside of the engine itself, it is the Name if not defined.
	// +optional
	NameInEngine string `json:"nameInEngine,omitempty"`

	// The model type, which is used to describe what the model is used for.
	// The default type is Text2Text, there are maybe more, and combination on the multi modal models.
	// It does have much influence on the LLM serving actually because it all depends on other configurations to start the model serving
	// This is used mainly for an indicator.
	// +kubebuilder:default:=Text2Text
	// +optional
	ModelType string `json:"modelType,omitempty"`

	// The URL for the model to download from if the model is managed by a 3rd party storage like a webserver or S3 storage
	// In engine like Ollama, the url is not needed, because it uses `ollama pull nameInEngine` to do the download job.
	// +optional
	ModelUrl string `json:"modelUrl,omitempty"`

	// The inline ModelDeploymentSpec
	*ModelDeploymentSpec `json:",inline"`
}

// ModelDeploymentSpec defines a spec for LLM Model serving
// It maps to a Deployment resource and the Pod resources it manages.
type ModelDeploymentSpec struct {

	// Args specifies the arguments to pass to the app container.
	// +optional
	Args *[]string `json:"args,omitempty"`

	// Storage specifies where the models are found and loaded
	// and also the cache volume to speed up the LLM inference by some frameworks like vLLM
	// +optional
	Storage *LLMEngineStorage `json:"storage"`

	// Replicas specifies how many replicas the LLMEngine will be deployed.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Envs specifies the environment variables to pass to the container.
	// There maybe extra environment set by the init containers to be merged with
	// +optional
	Envs *[]corev1.EnvVar `json:"env,omitempty"`

	// The image used in the init container to prepare the LLM model from remote repository
	// It will be the app container image if not set.
	// +optional
	DownloadImage string `json:"downloadImage,omitempty"`

	// The scripts used to prepare the LLM model from the remote repository
	// The scripts content will be saved to a config map and mounted by the init container to execute
	// It supports substitutions when storing to the config map with some builtin variables
	// The variables are for model directory, model name in the engine, etc.
	// +optional
	DownloadScripts string `json:"downloadScripts,omitempty"`
}

type LLMEngineDeploymentSpec struct {

	// HTTPPort specifies the open HTTP port for the engine inside of the container
	// It applies to all LLM models managed by this engine
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=8080
	// +optional
	HTTPPort int32 `json:"httpPort,omitempty"`

	// ServicePort specifies the port for the ClusterIP Service managed by this engine.
	// It applies to all LLM models managed by this engine
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default:=8080
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`

	// Image specifies the container image to use for the engine.
	// It applies to all LLM models managed by this engine
	// +optional
	Image string `json:"image"`
}

// LLMEngineSpec defines the desired state of LLMEngine.
type LLMEngineSpec struct {
	// Type specifies the type of LLM engine (e.g., ollama).
	// Each type corresponds a different default configuration set.
	// +kubebuilder:validation:Required
	EngineType *LLMEngineType `json:"engineType"`

	// The deployment runtime related settings, split it out to group with ModelDeploymentSpec without the whole LLMEngineSpec involved.
	*LLMEngineDeploymentSpec `json:",inline"`

	// The default deployment spec managed by this engine which applies to all LLM models managed by this engine
	// Each LLM model can override each configuration
	// It is optional, and there is a builtin configuration for each engine type.
	// The precedences of the set up are:  model config > engine default > builtin default
	// +optional
	*ModelDeploymentSpec `json:",inline"`

	// The LLM Models to be served by this engine
	// Each model corresponds a Deployment resource and a Service for the deployment
	// Each Deployment resource may have multiple pods according to replicas value
	// Each Service is a ClusterIP service which has a unique name composed by engine type and model name
	// All Deployment and Service resources are managed by this Operator
	// Changes to the whole LLMEngine resource may not trigger the recreation of the Deployment and Service if there is no change for it.
	// If no models are defined, it only starts the engine without any LLM models ready for the inference services.
	// +optional
	Models *[]ModelSpec `json:"models,omitempty"`
}

// LLMEngineStatus defines the observed state of LLMEngine.
// TODO: may add status of the models it manages ?
type LLMEngineStatus struct {
	// Conditions represent the latest available observations of the LLMEngine's state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
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
