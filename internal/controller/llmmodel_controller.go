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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
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
		if client.IgnoreNotFound(err) == nil {
			logger.Error(err, "Failed to get LLMEngine",
				"engineRef", engineRef, "namespace", req.Namespace)
			// Set failure condition on LLMModel status
			condition := metav1.Condition{
				Type:    "Ready",
				Status:  metav1.ConditionFalse,
				Reason:  "EngineNotFound",
				Message: fmt.Sprintf("No LLMEngine found with name=%s", engineRef),
			}
			if err := r.updateLLMModelStatus(ctx, req, &condition); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Set ownerReference to the engine
	if err := ctrl.SetControllerReference(llmEngine, llmModel, r.Scheme); err != nil {
		logger.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	if err := r.reconcileLLMModel(ctx, req, llmEngine, llmModel); err != nil {
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

	return ctrl.Result{}, nil
}

func (r *LLMModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aitrigramv1.LLMModel{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("llmmodel").
		Complete(r)
}
