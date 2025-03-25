package controller

import (
	corev1 "k8s.io/api/core/v1"
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

// Extra spaces using a map
func MergeEnvs(envs ...*[]corev1.EnvVar) *[]corev1.EnvVar {
	mm := make(map[string]corev1.EnvVar)
	for _, m := range envs {
		for _, mi := range *m {
			// the later one overrieds the previous ones
			mm[mi.Name] = mi
		}
	}
	result := make([]corev1.EnvVar, 0, len(mm))
	for _, v := range mm {
		result = append(result, v)
	}
	return &result
}
