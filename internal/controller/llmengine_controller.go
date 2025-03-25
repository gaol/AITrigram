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
	"context"
	"errors"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

// LLMEngineReconciler reconciles a LLMEngine object
type LLMEngineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop whtimeich aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LLMEngine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *LLMEngineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the LLMEngine instance
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, req.NamespacedName, llmEngine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// check status
	if len(llmEngine.Status.Conditions) == 0 {
		condition := &metav1.Condition{
			Type:    "Available",
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation"}
		if err := r.updateLLMEngineStatus(ctx, req, condition); err != nil {
			logger.Error(err, "Failed to update LLMEngine Status")
			return ctrl.Result{}, err
		}
	}
	defaultSpec := DefaultLLMEngineSpec(llmEngine.Spec.EngineType)
	// Check if the deployment already exists
	// Create a new deployment according to the spec.
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		// Failed to get the deployment info, but maybe because of not found
		if apierrors.IsNotFound(err) {
			// create one, and return
			newDeployment, err := r.newLLMEngineDeployment(llmEngine, defaultSpec)
			if err != nil {
				logger.Error(err, "Failed to define new Deployment resource for LLMEngine")
				// update status to false
				condition := &metav1.Condition{
					Type:    "Available",
					Status:  metav1.ConditionFalse,
					Reason:  "Reconciling",
					Message: fmt.Sprintf("Failed to create deployment for resource of LLMEngine: (%s): (%s)", llmEngine.Name, err),
				}
				if err := r.updateLLMEngineStatus(ctx, req, condition); err != nil {
					logger.Error(err, "Failed to update LLMEngine status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			logger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			if err := r.Create(ctx, newDeployment); err != nil {
				logger.Error(err, "Failed to create the new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
				return ctrl.Result{}, err
			}
			// deployment created successfully, requeue it in 10 seconds
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get the Deployment for LLMEngine")
		return ctrl.Result{}, err
	}
	//OK, the deployment works fine,
	//we need to sync the replica in case someone update the replicas of the deployment
	size := *llmEngine.Spec.Replicas
	if *deployment.Spec.Replicas != size {
		deployment.Spec.Replicas = &size
		if err := r.Update(ctx, deployment); err != nil {
			logger.Error(err, "Failed to update Deployment",
				"Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			condition := &metav1.Condition{
				Type:    "Available",
				Status:  metav1.ConditionFalse,
				Reason:  "Resizing",
				Message: fmt.Sprintf("Failed to update the size for the custom resource (%s): (%s)", llmEngine.Name, err)}

			if err := r.updateLLMEngineStatus(ctx, req, condition); err != nil {
				logger.Error(err, "Failed to update LLMEngine status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	// create service for each deployment
	llmService := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, llmService); err != nil {
		// not found yet
		if apierrors.IsNotFound(err) {
			newService, err := r.newLLMEngineService(llmEngine, defaultSpec)
			if err != nil {
				logger.Error(err, "Failed to create Service Resource for LLMEngine")
				return ctrl.Result{}, err
			}
			logger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			if err := r.Create(ctx, newService); err != nil {
				logger.Error(err, "Failed to create a new Service for LLMEngine")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get the service for LLMEngine")
		return ctrl.Result{}, err
	}
	// The following implementation will update the status
	found := false
	for _, item := range llmEngine.Status.Conditions {
		if item.Type == "Available" && item.Status == metav1.ConditionTrue {
			found = true
			break
		}
	}
	if !found {
		condition := &metav1.Condition{
			Type:    "Available",
			Status:  metav1.ConditionTrue,
			Reason:  "Reconciling",
			Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", llmEngine.Name, llmEngine.Spec.Replicas)}

		if err := r.updateLLMEngineStatus(ctx, req, condition); err != nil {
			logger.Error(err, "Failed to update LLMEngine Status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *LLMEngineReconciler) updateLLMEngineStatus(ctx context.Context, req ctrl.Request, condition *metav1.Condition) error {
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, req.NamespacedName, llmEngine); err != nil {
		return client.IgnoreNotFound(err)
	}
	if reflect.DeepEqual(llmEngine.Status.Conditions, condition) {

	}
	meta.SetStatusCondition(&llmEngine.Status.Conditions, *condition)
	if err := r.Status().Update(ctx, llmEngine); err != nil {
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

func (r *LLMEngineReconciler) newLLMEngineDeployment(llmEngine *aitrigramv1.LLMEngine, defaultSpec *aitrigramv1.LLMEngineSpec) (*appsv1.Deployment, error) {
	replicas := llmEngine.Spec.Replicas
	engineType := llmEngine.Spec.EngineType
	image := llmEngine.Spec.Image
	httpPort := llmEngine.Spec.HTTPPort
	args := llmEngine.Spec.Args
	storage := &llmEngine.Spec.Stroage
	if storage == nil {
		return nil, errors.New(fmt.Sprintf("No storage defined for engineType: %s, name: %s", string(*engineType), llmEngine.Name))
	}
	if image == nil {
		image = defaultSpec.Image
	}
	if image == nil {
		return nil, errors.New(fmt.Sprintf("No image defined for engineType: %s, name: %s", string(*engineType), llmEngine.Name))
	}
	if httpPort == nil {
		httpPort = defaultSpec.HTTPPort
	}
	if args == nil {
		args = defaultSpec.Args
	}
	volumes, volumeMounts := CacheAndModelsMount(&llmEngine.Spec)
	appLables := map[string]string{"app": "aitrigram-llmengine", "instance": llmEngine.Name}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      llmEngine.Name,
			Namespace: llmEngine.Namespace,
			Labels:    appLables,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: appLables,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appLables,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Volumes: volumes,
					//TODO: add NodeSelector ? ENV ? OLLAMA_MODELS path,
					Containers: []corev1.Container{{
						Image:           *image,
						Name:            string(*engineType) + "-" + llmEngine.Name,
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             ptr.To(true),
							RunAsUser:                ptr.To(int64(1001)),
							AllowPrivilegeEscalation: ptr.To(false),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: *httpPort,
							Name:          "http",
						}},
						Command:      *args,
						VolumeMounts: volumeMounts,
						// Env: []corev1.EnvVar{{
						// 	Name:  "OLLAMA_MODELS",
						// 	Value: llmEngine.Spec.ModelPath,
						// }},
					}},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(llmEngine, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMEngineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aitrigramv1.LLMEngine{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("llmengine").
		Complete(r)
}
