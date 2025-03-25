package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func (r *LLMEngineReconciler) reconcileLLMService(ctx context.Context, req ctrl.Request, llmEngine *aitrigramv1.LLMEngine) error {
	// create service for each deployment
	logger := log.FromContext(ctx)
	llmService := &corev1.Service{}
	defaultSpec := DefaultLLMEngineSpec(llmEngine.Spec.EngineType)
	if err := r.Get(ctx, req.NamespacedName, llmService); err != nil {
		// not found yet
		if apierrors.IsNotFound(err) {
			newService, err := r.newLLMEngineService(llmEngine, defaultSpec)
			if err != nil {
				logger.Error(err, "Failed to create Service Resource for LLMEngine")
				return err
			}
			logger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			if err := r.Create(ctx, newService); err != nil {
				logger.Error(err, "Failed to create a new Service for LLMEngine")
				return err
			}
			return nil
		}
		logger.Error(err, "Failed to get the service for LLMEngine")
		return err
	}
	return nil
}

func (r *LLMEngineReconciler) newLLMEngineService(llmEngine *aitrigramv1.LLMEngine, defaultSpec *aitrigramv1.LLMEngineSpec) (*corev1.Service, error) {
	appLables := map[string]string{"app": "aitrigram-llmengine", "instance": llmEngine.Name}
	httpPort := llmEngine.Spec.HTTPPort
	if httpPort == nil {
		httpPort = defaultSpec.HTTPPort
	}
	servicePort := llmEngine.Spec.ServicePort
	if servicePort == nil {
		servicePort = defaultSpec.ServicePort
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llmEngine.Name,
			Namespace: llmEngine.Namespace,
			Labels:    appLables,
		},
		Spec: corev1.ServiceSpec{
			Selector: appLables,
			Ports: []corev1.ServicePort{
				{
					Port:       *servicePort,
					TargetPort: intstr.FromInt32(*httpPort),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	// Set the ownerRef for the Service
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(llmEngine, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil
}
