package controller

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type DownloadScriptsTemplate struct {
	ModelName string
	ModelUrl  string
	ModelDir  string
}

// Reconcile the deployment for a LLM model
// when this method is called, all values have been recalculated with consideration of default values and user desired changes.
func (r *LLMEngineReconciler) reconcileLLMDeployment(ctx context.Context, req ctrl.Request, deploymentParams ReconcileParams) error {

	logger := log.FromContext(ctx)

	deploymentName := strings.ToLower(string(*deploymentParams.engineType) + "-" + strings.ReplaceAll(deploymentParams.modelSpec.Name, ".", "-"))
	deployment := &appsv1.Deployment{}
	nameSpaceName := &types.NamespacedName{
		Namespace: req.Namespace,
		Name:      deploymentName,
	}
	err := r.Get(ctx, *nameSpaceName, deployment)
	if err != nil {
		// Failed to get the deployment info, but maybe because of not found
		if apierrors.IsNotFound(err) {
			// create one, and return
			newDeployment, err := r.newLLMEngineDeployment(nameSpaceName, deploymentParams)
			if err != nil {
				logger.Error(err, "Failed to define new Deployment resource for LLMEngine")
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
	// Now the deployment has been created, but maybe need to update, let's calculate it
	existingCopy := deployment.DeepCopy()
	desired, err := r.newLLMEngineDeployment(nameSpaceName, deploymentParams)
	if err != nil {
		logger.Error(err, "Failed to define new Deployment resource for LLMEngine")
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
		logger.Info("Deployment is already up-to-date")
		return nil
	}
	if err := r.Client.Patch(ctx, deployment, client.MergeFrom(desired)); err != nil {
		logger.Error(err, "Failed to update the deployment")
		return err
	}
	return nil
}

func generateInitScript(scripts string, data DownloadScriptsTemplate) (string, error) {
	if scripts == "" {
		return scripts, nil
	}
	template, err := template.New("initScript").Parse(scripts)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := template.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

// The new deployment has the ownerReferences to the llmEngine CR, so it will be handled automatically by the core
func (r *LLMEngineReconciler) newLLMEngineDeployment(nameSpaceName *types.NamespacedName, deploymentParams ReconcileParams) (*appsv1.Deployment, error) {
	replicas := deploymentParams.modelSpec.Replicas
	image := deploymentParams.engineDeploymentSpec.Image
	httpPort := deploymentParams.engineDeploymentSpec.HTTPPort
	args := deploymentParams.modelSpec.Args
	envs := deploymentParams.modelSpec.Envs
	volumes, volumeMounts := cacheAndModelsMount(deploymentParams.modelSpec)
	appLabels := map[string]string{"app": "aitrigram-llmengine", "instance": nameSpaceName.Name}
	downloadScriptsTemplate := DownloadScriptsTemplate{
		ModelName: deploymentParams.modelSpec.NameInEngine,
		ModelUrl:  deploymentParams.modelSpec.ModelUrl,
		ModelDir:  deploymentParams.modelSpec.Storage.ModelsStorage.Path,
	}
	downloadScripts, err := generateInitScript(deploymentParams.modelSpec.DownloadScripts, downloadScriptsTemplate)
	if err != nil {
		return nil, err
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameSpaceName.Name,
			Namespace: nameSpaceName.Namespace,
			Labels:    appLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: appLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: volumes,
					InitContainers: []corev1.Container{{
						Image:        deploymentParams.modelSpec.DownloadImage,
						Name:         "init-" + nameSpaceName.Name,
						VolumeMounts: volumeMounts,
						Env:          *envs,
						Command:      []string{"/bin/sh", "-c"},
						Args:         []string{downloadScripts},
					}},
					Containers: []corev1.Container{{
						Image:           image,
						Name:            nameSpaceName.Name,
						ImagePullPolicy: corev1.PullIfNotPresent,

						Ports: []corev1.ContainerPort{{
							ContainerPort: httpPort,
							Name:          "http",
						}},
						Command:      *args,
						VolumeMounts: volumeMounts,
						Env:          *envs,
					}},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(deploymentParams.llmEngine, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}
