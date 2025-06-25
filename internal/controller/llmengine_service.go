package controller

import (
	"context"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Each LLM Model matches a Service for it, which exposes to the cluster, and makes it possible for LB from external cluster
func (r *LLMModelReconciler) reconcileLLMService(ctx context.Context, req ctrl.Request, serviceParams ReconcileParams) error {
	// create service for each deployment
	logger := log.FromContext(ctx)

	serviceName := strings.ToLower(string(*&serviceParams.llmEngine.Spec.EngineType) + "-" + strings.ReplaceAll(serviceParams.model.Spec.Name, ".", "-"))
	nameSpaceName := &types.NamespacedName{
		Namespace: req.Namespace,
		Name:      serviceName,
	}

	llmService := &corev1.Service{}
	if err := r.Get(ctx, *nameSpaceName, llmService); err != nil {
		// not found yet
		if apierrors.IsNotFound(err) {
			newService, err := r.newLLMEngineService(nameSpaceName, serviceParams)
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
	// check the update
	existingCopy := llmService.DeepCopy()
	desired, err := r.newLLMEngineService(nameSpaceName, serviceParams)
	if err != nil {
		logger.Error(err, "Failed to define new Service resource for LLMEngine")
		return err
	}
	// make a deep copy and ignore some fields for comparison
	desiredCopy := desired.DeepCopy()
	desiredCopy.ObjectMeta.OwnerReferences = nil
	desiredCopy.ObjectMeta.ResourceVersion = ""
	existingCopy.ObjectMeta.ResourceVersion = ""
	existingCopy.ObjectMeta.OwnerReferences = nil
	existingCopy.Spec = desired.Spec
	if reflect.DeepEqual(existingCopy.Spec, desiredCopy.Spec) {
		logger.Info("Service is already up-to-date")
		return nil
	}
	if err := r.Client.Patch(ctx, llmService, client.MergeFrom(desired)); err != nil {
		logger.Error(err, "Failed to update the service")
		return err
	}

	return nil
}

func (r *LLMModelReconciler) newLLMEngineService(nameSpaceName *types.NamespacedName, serviceParams ReconcileParams) (*corev1.Service, error) {
	appLabels := map[string]string{"app": "aitrigram-llmmodel", "instance": nameSpaceName.Name}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameSpaceName.Name,
			Namespace: nameSpaceName.Namespace,
			Labels:    appLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: appLabels,
			Ports: []corev1.ServicePort{
				{
					Port:       serviceParams.llmEngine.Spec.ServicePort,
					TargetPort: intstr.FromInt32(serviceParams.llmEngine.Spec.Port),
				},
			},
			Type:            corev1.ServiceTypeClusterIP,
			SessionAffinity: corev1.ServiceAffinityClientIP,
		},
	}
	// Set the ownerRef for the Service
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(serviceParams.model, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil
}
