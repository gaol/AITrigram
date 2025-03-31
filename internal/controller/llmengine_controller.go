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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LLMEngine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// No matter if the call was triggered by changes in the owned resources or LLMEngine itself,
// the ctx.Get() returns the LLMEngine, not the sub resources like Deployment and Service.
func (r *LLMEngineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Fetch the LLMEngine instance
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, req.NamespacedName, llmEngine); err != nil {
		// if it is a not found error, it has been deleted already.
		// All sub resources have the ownership with this CRD, so we don't need to worry about the deletion.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// llmEngine is not nil from now, the instance is the the modified version that will will take effect.

	// The engine scope changes won't lead to reconcile again, the changes will be applied to the models if
	// any runtime changes were made after recalculating the models.
	// So let's reconcile LLM models

	modelSpecs := llmEngine.Spec.Models
	if modelSpecs != nil && len(*modelSpecs) > 0 {
		// there are some LLM defined, let's create resources for each of them
		// without models, no engine starts
		for _, mSpec := range *modelSpecs {
			if err := r.reconcileLLMModel(ctx, req, llmEngine, &mSpec); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
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
