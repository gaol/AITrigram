package controller

import (
	"context"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ReconcileParams struct {
	llmEngine *aitrigramv1.LLMEngine
	// the ModelDeployment in the model has taken the values in the llmEngine
	model *aitrigramv1.LLMModel
}

func (r *LLMModelReconciler) reconcileLLMModel(ctx context.Context, req ctrl.Request,
	llmEngine *aitrigramv1.LLMEngine, model *aitrigramv1.LLMModel) error {

	// llmEngine has now the all values set because it is retrieved from the cluster
	// the same for the llmEngine.Spec
	modelDeploymentSpec, err := MergeModelDeploymentTemplate(llmEngine.Spec.ModelDeploymentTemplate, model.Spec.ModelDeployment)
	if err != nil {
		return err
	}
	model.Spec.ModelDeployment = modelDeploymentSpec

	params := ReconcileParams{
		llmEngine: llmEngine,
		model:     model,
	}
	if err := r.reconcileLLMDeployment(ctx, req, params); err != nil {
		return err
	}
	// reconcile service for this model
	if err := r.reconcileLLMService(ctx, req, params); err != nil {
		return err
	}
	return nil
}
