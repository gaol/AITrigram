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
	// Reconcile the associated deployment for LLMEngine
	// The new created deployment has ownReference to the LLMEngine
	if err := r.reconcileLLMDeployment(ctx, req, llmEngine); err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile the LLMService associated to the LLMEngine
	// The new created Serice has ownReference to this LLMEngine
	if err := r.reconcileLLMService(ctx, req, llmEngine); err != nil {
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

		logger.Info("Update LLMEngine Status to True")
		if err := r.updateLLMEngineStatus(ctx, req, condition); err != nil {
			logger.Error(err, "Failed to update LLMEngine Status")
			return ctrl.Result{}, err
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
