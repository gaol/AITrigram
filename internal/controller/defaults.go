package controller

import (
	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func cacheAndModelsMount(modelSpec *aitrigramv1.ModelSpec) ([]corev1.Volume, []corev1.VolumeMount) {
	// storage may be nil
	storage := modelSpec.Storage
	modelStorage := modelSpec.Storage.ModelsStorage
	var cacheStorage *aitrigramv1.CacheStorage
	if storage != nil {
		if storage.ModelsStorage != nil {
			modelStorage = storage.ModelsStorage
		}
		cacheStorage = storage.CacheStorage
	}
	modelVolume := corev1.Volume{
		Name:         "models",
		VolumeSource: modelStorage.VolumeSource,
	}
	modelVolumeMount := corev1.VolumeMount{
		Name:      "models",
		MountPath: modelStorage.Path,
	}
	if cacheStorage != nil {
		cacheVolume := corev1.Volume{
			Name: "cache",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: cacheStorage.EmptyDirVolumeSource,
			},
		}
		cacheVolumeMount := corev1.VolumeMount{
			Name:      "cache",
			MountPath: cacheStorage.Path,
		}
		return []corev1.Volume{modelVolume, cacheVolume}, []corev1.VolumeMount{modelVolumeMount, cacheVolumeMount}
	}
	return []corev1.Volume{modelVolume}, []corev1.VolumeMount{modelVolumeMount}
}

func defaultLLMEngineSpec(engineType *aitrigramv1.LLMEngineType) *aitrigramv1.LLMEngineSpec {
	if *engineType == aitrigramv1.LLMEngineTypeOllama {
		return defaultsOfOllamaEngine()
	}
	//TODO add more for other types
	return &aitrigramv1.LLMEngineSpec{}
}

// Returns default setup for Ollama engine
func defaultsOfOllamaEngine() *aitrigramv1.LLMEngineSpec {
	ollama := aitrigramv1.LLMEngineTypeOllama
	ollamaEngine := &aitrigramv1.LLMEngineSpec{
		EngineType:              &ollama,
		LLMEngineDeploymentSpec: defaultsLLMDeploymentSpecOllama(),
		ModelDeploymentSpec:     defaultsModelDeploymentSpecOllama(),
	}
	return ollamaEngine
}

const (
	defaultOllamaImage string = "ollama/ollama:latest"
)

func defaultsLLMDeploymentSpecOllama() *aitrigramv1.LLMEngineDeploymentSpec {
	return &aitrigramv1.LLMEngineDeploymentSpec{
		Image:       defaultOllamaImage,
		HTTPPort:    11434,
		ServicePort: 8080,
	}
}

func defaultsModelDeploymentSpecOllama() *aitrigramv1.ModelDeploymentSpec {
	cacheSizeLimit := resource.MustParse("2Gi")
	return &aitrigramv1.ModelDeploymentSpec{
		Args:            &[]string{"/bin/ollama", "serve"},
		Replicas:        1,
		DownloadImage:   defaultOllamaImage,
		DownloadScripts: `ollama serve & sleep 10 && ollama pull {{ .ModelName }}`,
		Storage: &aitrigramv1.LLMEngineStorage{
			ModelsStorage: &aitrigramv1.ModelStorage{
				Path: "/models",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			CacheStorage: &aitrigramv1.CacheStorage{
				Path: "/cache_dir",
				EmptyDirVolumeSource: &corev1.EmptyDirVolumeSource{
					SizeLimit: &cacheSizeLimit,
				},
			},
		},
		Envs: &[]corev1.EnvVar{
			{
				Name:  "OLLAMA_MODELS",
				Value: "/models",
			},
			{
				Name:  "OLLAMA_CACHE_DIR",
				Value: "/cache_dir",
			},
		},
	}
}

var (
	ollamaEngineType                                   = aitrigramv1.LLMEngineTypeOllama
	DefaultOllamaEngineSpec *aitrigramv1.LLMEngineSpec = defaultLLMEngineSpec(&ollamaEngineType)
)

// Merge the ModelDeploymentSpecs, the later settings overrides the previous ones
// So make sure the ones you want to keep in the last arguments.
func mergeModelDeploymentSpecs(modelSpecs ...*aitrigramv1.ModelDeploymentSpec) (*aitrigramv1.ModelDeploymentSpec, error) {
	result := &aitrigramv1.ModelDeploymentSpec{}
	for _, ms := range modelSpecs {
		if ms == nil {
			continue
		}
		if ms.Args != nil {
			result.Args = ms.Args
		}
		if ms.DownloadImage != "" {
			result.DownloadImage = ms.DownloadImage
		}
		if ms.DownloadScripts != "" {
			result.DownloadScripts = ms.DownloadScripts
		}
		if ms.Replicas != 0 {
			result.Replicas = ms.Replicas
		}
		if ms.Envs != nil {
			envs, err := MergeSliceByName(result.Envs, ms.Envs)
			if err != nil {
				return nil, err
			}
			result.Envs = envs
		}
		if ms.Storage != nil {
			result.Storage = mergeStorages(result.Storage, ms.Storage)
		}
	}
	return result, nil
}

// Merge the LLMEngineDeploymentSpec, the later overrides the previous ones
func mergeLLMDeploymentSpecs(llmDeploymentSpecs ...*aitrigramv1.LLMEngineDeploymentSpec) *aitrigramv1.LLMEngineDeploymentSpec {
	result := &aitrigramv1.LLMEngineDeploymentSpec{}
	for _, llmDeploymentSpec := range llmDeploymentSpecs {
		if llmDeploymentSpec == nil {
			continue
		}
		if llmDeploymentSpec.HTTPPort != 0 {
			result.HTTPPort = llmDeploymentSpec.HTTPPort
		}
		if llmDeploymentSpec.Image != "" {
			result.Image = llmDeploymentSpec.Image
		}
		if llmDeploymentSpec.ServicePort != 0 {
			result.ServicePort = llmDeploymentSpec.ServicePort
		}
	}
	return result
}

func mergeStorages(storages ...*aitrigramv1.LLMEngineStorage) *aitrigramv1.LLMEngineStorage {
	result := &aitrigramv1.LLMEngineStorage{}
	for _, storage := range storages {
		if storage == nil {
			continue
		}
		if storage.CacheStorage != nil {
			result.CacheStorage = storage.CacheStorage
		}
		if storage.ModelsStorage != nil {
			result.ModelsStorage = storage.ModelsStorage
		}

	}
	return result
}
