package controller

import (
	"context"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func (r *LLMEngineReconciler) reconcileLLMDeployment(ctx context.Context, req ctrl.Request, llmEngine *aitrigramv1.LLMEngine) error {
	logger := log.FromContext(ctx)
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
					return err
				}
				return err
			}
			logger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			if err := r.Create(ctx, newDeployment); err != nil {
				logger.Error(err, "Failed to create the new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
				return err
			}
			// deployment created successfully, requeue it in 10 seconds
			return nil
		}
		logger.Error(err, "Failed to get the Deployment for LLMEngine")
		return err
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
				return err
			}
			return err
		}
	}
	return nil
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
					// TODO cover the security in the future
					// SecurityContext: &corev1.PodSecurityContext{
					// 	RunAsNonRoot: ptr.To(true),
					// 	SeccompProfile: &corev1.SeccompProfile{
					// 		Type: corev1.SeccompProfileTypeRuntimeDefault,
					// 	},
					// },
					Volumes: volumes,
					//TODO: add NodeSelector ? ENV ? OLLAMA_MODELS path,
					Containers: []corev1.Container{{
						Image:           *image,
						Name:            string(*engineType) + "-" + llmEngine.Name,
						ImagePullPolicy: corev1.PullIfNotPresent,
						// Ensure restrictive context for the container
						// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
						// SecurityContext: &corev1.SecurityContext{
						// 	RunAsNonRoot:             ptr.To(true),
						// 	RunAsUser:                ptr.To(int64(1001)),
						// 	AllowPrivilegeEscalation: ptr.To(false),
						// 	Capabilities: &corev1.Capabilities{
						// 		Drop: []corev1.Capability{
						// 			"ALL",
						// 		},
						// 	},
						// },
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
