package controller

import (
	"fmt"
	"reflect"
	"slices"

	corev1 "k8s.io/api/core/v1"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// The merge uses extra spaces using a map
func MergeSliceByName[O interface{}](objs ...*[]O) (*[]O, error) {
	mm := make(map[string]O)
	for _, o := range objs {
		if o == nil {
			continue
		}
		for _, mi := range *o {
			// the later one overrides the previous ones
			name, r := GetFieldValue(mi, "Name")
			if r {
				mm[name.(string)] = mi
			} else {
				return nil, fmt.Errorf("there is no Name field in: %v", mi)
			}
		}
	}
	result := make([]O, 0, len(mm))
	for _, v := range mm {
		result = append(result, v)
	}
	return &result, nil
}

func GetFieldValue(obj interface{}, fieldName string) (interface{}, bool) {
	val := reflect.ValueOf(obj)

	// If it's a pointer, get the element it points to
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Must be a struct to have fields
	if val.Kind() != reflect.Struct {
		return nil, false
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}

	return field.Interface(), true
}

func LLMEngineSpecEquals(spec1, spec2 *aitrigramv1.LLMEngineSpec) bool {
	if spec1 == nil || spec2 == nil {
		return spec1 == spec2
	}

	if spec1.EngineType != spec2.EngineType {
		return false
	}
	if spec1.Image != spec2.Image {
		return false
	}
	if spec1.Port != spec2.Port {
		return false
	}
	if spec1.ServicePort != spec2.ServicePort {
		return false
	}
	if spec1.ModelDeploymentTemplate != nil && spec2.ModelDeploymentTemplate != nil {
		return ModelDeploymentEquals(spec1.ModelDeploymentTemplate, spec2.ModelDeploymentTemplate)
	} else if (spec1.ModelDeploymentTemplate) != nil != (spec2.ModelDeploymentTemplate != nil) {
		return false
	}
	return true
}

func ModelDeploymentEquals(dep1, dep2 *aitrigramv1.ModelDeploymentTemplate) bool {
	if dep1 == nil || dep2 == nil {
		return dep1 == dep2
	}
	if dep1.DownloadImage != dep2.DownloadImage {
		return false
	}
	if dep1.DownloadScripts != dep2.DownloadScripts {
		return false
	}
	if !slices.Equal(dep1.Args, dep2.Args) {
		return false
	}
	if !envsEquals(*dep1.Envs, *dep2.Envs) {
		return false
	}
	if !reflect.DeepEqual(dep1.Storage, dep2.Storage) {
		return false
	}

	return true
}

func envEquals(env1, env2 corev1.EnvVar) bool {
	if env1.Name != env2.Name {
		return false
	}
	if env1.Value != env2.Value {
		return false
	}
	return true
}

func envsEquals(env1, env2 []corev1.EnvVar) bool {
	if len(env1) != len(env2) {
		return false
	}
	for i := range env1 {
		if !envInSlice(env1[i], env2) {
			return false
		}
	}
	return true
}

func envInSlice(env corev1.EnvVar, envs []corev1.EnvVar) bool {
	for i := range envs {
		if envEquals(env, envs[i]) {
			return true
		}
	}
	return false
}
