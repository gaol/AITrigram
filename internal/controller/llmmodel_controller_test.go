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

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func Test_LLMModelDefault(t *testing.T) {
	t.Parallel()
	// ollamaEngineType := aitrigramv1.LLMEngineTypeOllama
	// ollamaDefaultEngineSpec := *DefaultLLMEngineSpec(&ollamaEngineType)
	// cacheSizeLimit := resource.MustParse("2Gi")

	cases := map[string]struct {
		llmEngineSpec aitrigramv1.LLMEngineSpec
		llmModelSpec  aitrigramv1.LLMModelSpec
		expected      aitrigramv1.ModelDeploymentTemplate
	}{
		"default-empty": {
			llmEngineSpec: aitrigramv1.LLMEngineSpec{
				EngineType: aitrigramv1.LLMEngineTypeOllama,
				ModelDeploymentTemplate: &aitrigramv1.ModelDeploymentTemplate{
					Args:          []string{"/bin/ollama", "serve", "-b", "127.0.0.1"},
					DownloadImage: "ollama/ollama:latest",
					Envs: &[]corev1.EnvVar{
						{
							Name:  "OLLAMA_MODELS",
							Value: "/models",
						},
						{
							Name:  "OLLAMA_CACHE_DIR",
							Value: "/cache_dir_old",
						},
					},
				},
			},
			llmModelSpec: aitrigramv1.LLMModelSpec{
				Name:      "test-model",
				EngineRef: "test-engine",
				Replicas:  1,
				ModelDeployment: &aitrigramv1.ModelDeploymentTemplate{
					Args: []string{"/bin/ollama", "serve", "-b", "0.0.0.0"},
					Envs: &[]corev1.EnvVar{
						{
							Name:  "OLLAMA_MODELS",
							Value: "/new_models",
						},
						{
							Name:  "OLLAMA_CACHE_DIR",
							Value: "/cache_dir",
						},
						{
							Name:  "NEW_ENV",
							Value: "new env value",
						},
					},
				},
			},
			expected: aitrigramv1.ModelDeploymentTemplate{
				Args:          []string{"/bin/ollama", "serve", "-b", "0.0.0.0"},
				DownloadImage: "ollama/ollama:latest",
				Envs: &[]corev1.EnvVar{
					{
						Name:  "OLLAMA_MODELS",
						Value: "/new_models",
					},
					{
						Name:  "OLLAMA_CACHE_DIR",
						Value: "/cache_dir",
					},
					{
						Name:  "NEW_ENV",
						Value: "new env value",
					},
				},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result, err := MergeModelDeploymentTemplate(c.llmEngineSpec.ModelDeploymentTemplate, c.llmModelSpec.ModelDeployment)
			require.NoError(t, err)
			if !ModelDeploymentEquals(&c.expected, result) {
				t.Errorf("maps do not match.\nExpected: %#v\nActual: %#v", c.expected, *result)
			}
		})
	}
}
