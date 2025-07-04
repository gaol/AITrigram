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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

// LLMEngineReconciler reconciles a LLMEngine object
type LLMEngineReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	OperatorNamespace string
	OperatorPodName   string
}

// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aitrigram.ihomeland.cn,resources=llmengines/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

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
	logger := log.FromContext(ctx)
	// Fetch the LLMEngine instance
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, req.NamespacedName, llmEngine); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	desired, err := MergeLLMSpecs(DefaultLLMEngineSpec(&llmEngine.Spec.EngineType).DeepCopy(), &llmEngine.Spec)
	if err != nil {
		return ctrl.Result{}, nil
	}
	if LLMEngineSpecEquals(&llmEngine.Spec, desired) {
		logger.Info("LLMEngine is already up-to-date")
		return ctrl.Result{}, nil
	}
	llmEngine.Spec = *desired
	logger.Info("Update LLMEngine Spec")
	if err := r.Client.Update(ctx, llmEngine); err != nil {
		logger.Error(err, "Failed to update the llmengine")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *LLMEngineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aitrigramv1.LLMEngine{}).
		Owns(&aitrigramv1.LLMModel{}).
		Named("llmengine").
		Complete(r)
}
