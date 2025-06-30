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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// LLMModelReconciler reconciles a LLMModel object
type LLMModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmmodels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmmodels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmmodels/finalizers,verbs=update
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *LLMModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	llmModel := &aitrigramv1.LLMModel{}
	if err := r.Get(ctx, req.NamespacedName, llmModel); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	engineRef := llmModel.Spec.EngineRef
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, client.ObjectKey{Name: engineRef, Namespace: req.Namespace}, llmEngine); err != nil {
		if apierrors.IsNotFound(err) {
			if llmModel.GetDeletionTimestamp() != nil || llmEngine.GetDeletionTimestamp() != nil {
				// llmModel has been deleted, just ignore
				logger.Info("llmEngine has been deleted, llmModel will be deleted too, ignore it.")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Failed to get LLMEngine",
				"engineRef", engineRef, "namespace", req.Namespace)
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		return ctrl.Result{}, err
	}

	// llmEngine has now the all values set because it is retrieved from the cluster
	// the same for the llmEngine.Spec
	modelDeploymentSpec, err := MergeModelDeploymentTemplate(llmEngine.Spec.ModelDeploymentTemplate, llmModel.Spec.ModelDeployment)
	if err != nil {
		return ctrl.Result{}, err
	}
	// somehow, it is nil ...
	if modelDeploymentSpec == nil {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	if !ModelDeploymentEquals(llmModel.Spec.ModelDeployment, modelDeploymentSpec) {
		logger.Info("Updating the LLMModel")
		llmModel.Spec.ModelDeployment = modelDeploymentSpec
		if err := r.Client.Update(ctx, llmModel); err != nil {
			logger.Error(err, "Failed to update the llmengine")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	params := ReconcileParams{
		llmEngine: llmEngine,
		model:     llmModel,
	}
	if err := r.reconcileLLMDeployment(ctx, req, params); err != nil {
		return ctrl.Result{}, err
	}
	// reconcile service for this model
	if err := r.reconcileLLMService(ctx, req, params); err != nil {
		return ctrl.Result{}, err
	}

	// Set Ready condition if successful
	condition := metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionTrue,
		Reason:  "Deployed",
		Message: "LLMModel successfully deployed",
	}
	if err := r.updateLLMModelStatus(ctx, req, &condition); err != nil {
		return ctrl.Result{}, err
	}

	// Set ownerReference to the engine
	if err := ctrl.SetControllerReference(llmEngine, llmModel, r.Scheme); err != nil {
		logger.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type ReconcileParams struct {
	llmEngine *aitrigramv1.LLMEngine
	// the ModelDeployment in the model has taken the values in the llmEngine
	model *aitrigramv1.LLMModel
}

func (r *LLMModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aitrigramv1.LLMModel{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("llmmodel").
		Complete(r)
}
