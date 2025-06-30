package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
)

func (r *LLMEngineReconciler) UpdateLLMEngineStatus(ctx context.Context, req ctrl.Request, condition *metav1.Condition) error {
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

func (r *LLMModelReconciler) updateLLMModelStatus(ctx context.Context, req ctrl.Request, condition *metav1.Condition) error {
	llmModel := &aitrigramv1.LLMModel{}
	if err := r.Get(ctx, req.NamespacedName, llmModel); err != nil {
		return client.IgnoreNotFound(err)
	}
	meta.SetStatusCondition(&llmModel.Status.Conditions, *condition)
	if err := r.Status().Update(ctx, llmModel); err != nil {
		return err
	}
	return nil
}
