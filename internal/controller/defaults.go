package controller

import (
	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func CacheAndModelsMount(llmEngineSpec *aitrigramv1.LLMEngineSpec) ([]corev1.Volume, []corev1.VolumeMount) {
	// storage may be nil
	stroage := llmEngineSpec.Stroage
	modelStorage := &aitrigramv1.StroageMount{
		Path: "/models",
		VolumeSource: &corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	var cacheStorage *aitrigramv1.StroageMount
	if stroage != nil {
		if &stroage.ModelsStorage != nil {
			modelStorage = stroage.ModelsStorage
		}
		cacheStorage = stroage.CacheStorage
	}
	modelVolume := corev1.Volume{
		Name:         "models",
		VolumeSource: *modelStorage.VolumeSource,
	}
	modelVolumeMount := corev1.VolumeMount{
		Name:      "models",
		MountPath: modelStorage.Path,
	}
	if cacheStorage != nil {
		cacheVolume := corev1.Volume{
			Name:         "cache",
			VolumeSource: *cacheStorage.VolumeSource,
		}
		cacheVolumeMount := corev1.VolumeMount{
			Name:      "cache",
			MountPath: cacheStorage.Path,
		}
		return []corev1.Volume{modelVolume, cacheVolume}, []corev1.VolumeMount{modelVolumeMount, cacheVolumeMount}
	}
	return []corev1.Volume{modelVolume}, []corev1.VolumeMount{modelVolumeMount}
}

func DefaultLLMEngineSpec(engineType *aitrigramv1.LLMEngineType) *aitrigramv1.LLMEngineSpec {
	if *engineType == aitrigramv1.LLMEngineTypeOllama {
		return DefaultsOfOllamaEngine()
	}
	//TODO add more for other types
	return &aitrigramv1.LLMEngineSpec{}
}

// Returns default setup for Ollama engine
func DefaultsOfOllamaEngine() *aitrigramv1.LLMEngineSpec {
	image := "virt.lins-p1:5000/ollama/ollama:latest"
	var port int32 = 11434
	ollamaEngine := &aitrigramv1.LLMEngineSpec{
		Image:    &image,
		HTTPPort: &port,
		Args:     &[]string{"/bin/ollama", "serve"},
	}
	return ollamaEngine
}
