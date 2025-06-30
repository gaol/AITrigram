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

package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/require"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func Test_OllamaEngineDefault(t *testing.T) {
	t.Parallel()
	ollamaEngineType := aitrigramv1.LLMEngineTypeOllama
	ollamaDefaultEngineSpec := *DefaultLLMEngineSpec(&ollamaEngineType).DeepCopy()
	cacheSizeLimit := resource.MustParse("2Gi")

	cases := map[string]struct {
		llmEngineSpec aitrigramv1.LLMEngineSpec
		expected      aitrigramv1.LLMEngineSpec
	}{
		"default-empty": {
			llmEngineSpec: aitrigramv1.LLMEngineSpec{
				EngineType: aitrigramv1.LLMEngineTypeOllama,
			},
			expected: ollamaDefaultEngineSpec,
		},
		"default-empty-Engine": {
			llmEngineSpec: aitrigramv1.LLMEngineSpec{
				EngineType:  aitrigramv1.LLMEngineTypeOllama,
				Image:       "custom/ollama:latest",
				Port:        12345,
				ServicePort: 9090,
			},
			expected: aitrigramv1.LLMEngineSpec{
				EngineType:              aitrigramv1.LLMEngineTypeOllama,
				Image:                   "custom/ollama:latest",
				Port:                    12345,
				ServicePort:             9090,
				ModelDeploymentTemplate: ollamaDefaultEngineSpec.ModelDeploymentTemplate,
			},
		},
		"default-empty-model-template-envs": {
			llmEngineSpec: aitrigramv1.LLMEngineSpec{
				EngineType: aitrigramv1.LLMEngineTypeOllama,
				ModelDeploymentTemplate: &aitrigramv1.ModelDeploymentTemplate{
					Args:          []string{"/bin/ollama", "serve", "-b", "0.0.0.0"},
					DownloadImage: "ollama/ollama:download",
					Envs: &[]corev1.EnvVar{
						{
							Name:  "OLLAMA_MODELS",
							Value: "/new_models",
						},
						{
							Name:  "OLLAMA_CACHE_DIR",
							Value: "/cache_dir_new",
						},
					},
				},
			},
			expected: aitrigramv1.LLMEngineSpec{
				EngineType:  aitrigramv1.LLMEngineTypeOllama,
				Image:       ollamaDefaultEngineSpec.Image,
				Port:        ollamaDefaultEngineSpec.Port,
				ServicePort: ollamaDefaultEngineSpec.ServicePort,
				ModelDeploymentTemplate: &aitrigramv1.ModelDeploymentTemplate{
					Args:            []string{"/bin/ollama", "serve", "-b", "0.0.0.0"},
					DownloadImage:   "ollama/ollama:download",
					DownloadScripts: ollamaDefaultEngineSpec.ModelDeploymentTemplate.DownloadScripts,
					Storage:         ollamaDefaultEngineSpec.ModelDeploymentTemplate.Storage,
					Envs: &[]corev1.EnvVar{
						{
							Name:  "OLLAMA_MODELS",
							Value: "/new_models",
						},
						{
							Name:  "OLLAMA_CACHE_DIR",
							Value: "/cache_dir_new",
						},
					},
				},
			},
		},
		"default-empty-model-template-storage": {
			llmEngineSpec: aitrigramv1.LLMEngineSpec{
				EngineType: aitrigramv1.LLMEngineTypeOllama,
				ModelDeploymentTemplate: &aitrigramv1.ModelDeploymentTemplate{
					Storage: &aitrigramv1.LLMEngineStorage{
						ModelsStorage: &aitrigramv1.ModelStorage{
							Path: "/models_new",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						CacheStorage: &aitrigramv1.CacheStorage{
							Path: "/cache_dir_new",
							EmptyDirVolumeSource: &corev1.EmptyDirVolumeSource{
								SizeLimit: &cacheSizeLimit,
							},
						},
					},
				},
			},
			expected: aitrigramv1.LLMEngineSpec{
				EngineType:  aitrigramv1.LLMEngineTypeOllama,
				Image:       ollamaDefaultEngineSpec.Image,
				Port:        ollamaDefaultEngineSpec.Port,
				ServicePort: ollamaDefaultEngineSpec.ServicePort,
				ModelDeploymentTemplate: &aitrigramv1.ModelDeploymentTemplate{
					Args:            ollamaDefaultEngineSpec.ModelDeploymentTemplate.Args,
					DownloadImage:   ollamaDefaultEngineSpec.ModelDeploymentTemplate.DownloadImage,
					DownloadScripts: ollamaDefaultEngineSpec.ModelDeploymentTemplate.DownloadScripts,
					Storage: &aitrigramv1.LLMEngineStorage{
						ModelsStorage: &aitrigramv1.ModelStorage{
							Path: "/models_new",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						CacheStorage: &aitrigramv1.CacheStorage{
							Path: "/cache_dir_new",
							EmptyDirVolumeSource: &corev1.EmptyDirVolumeSource{
								SizeLimit: &cacheSizeLimit,
							},
						},
					},
					Envs: ollamaDefaultEngineSpec.ModelDeploymentTemplate.Envs,
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			engineType := c.llmEngineSpec.EngineType
			defaultSpec := DefaultLLMEngineSpec(&engineType)
			result, err := MergeLLMSpecs(defaultSpec, &c.llmEngineSpec)
			require.NoError(t, err)
			if !LLMEngineSpecEquals(&c.expected, result) {
				t.Errorf("maps do not match.\nExpected: %#v\nActual: %#v", c.expected, *result)
			}
		})
	}
}
