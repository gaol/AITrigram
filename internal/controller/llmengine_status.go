package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func (r *LLMEngineReconciler) updateLLMEngineStatus(ctx context.Context, req ctrl.Request, condition *metav1.Condition) error {
	llmEngine := &aitrigramv1.LLMEngine{}
	if err := r.Get(ctx, req.NamespacedName, llmEngine); err != nil {
		return client.IgnoreNotFound(err)
	}
	meta.SetStatusCondition(&llmEngine.Status.Conditions, *condition)
	if err := r.Status().Update(ctx, llmEngine); err != nil {
		return err
	}
	return nil
}
