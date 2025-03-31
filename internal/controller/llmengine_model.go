package controller

import (
	"context"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ReconcileParams struct {
	engineType           *aitrigramv1.LLMEngineType
	engineDeploymentSpec *aitrigramv1.LLMEngineDeploymentSpec
	llmEngine            *aitrigramv1.LLMEngine
	modelSpec            *aitrigramv1.ModelSpec
}

func (r *LLMEngineReconciler) reconcileLLMModel(ctx context.Context, req ctrl.Request,
	llmEngine *aitrigramv1.LLMEngine, modelSpec *aitrigramv1.ModelSpec) error {

	// default settings in the LLMEngineSpec
	// This needs to be calculated by merging the ones in modelSpec to the one from llmEngine.Spec.ModelDeploymentSpec
	// and it also takes consideration of the default settings by llmEngineType
	defaultSpec := defaultLLMEngineSpec(llmEngine.Spec.EngineType)

	modelDeploymentSpec, err := mergeModelDeploymentSpecs(defaultSpec.ModelDeploymentSpec, llmEngine.Spec.ModelDeploymentSpec, modelSpec.ModelDeploymentSpec)
	if err != nil {
		return err
	}
	modelSpec = modelSpec.DeepCopy()
	modelSpec.ModelDeploymentSpec = modelDeploymentSpec
	llmEngineDeploymentSpec := mergeLLMDeploymentSpecs(defaultSpec.LLMEngineDeploymentSpec, llmEngine.Spec.LLMEngineDeploymentSpec)

	// reconcile deployment for this model, if maybe multiple pods
	params := ReconcileParams{
		engineType:           llmEngine.Spec.EngineType,
		engineDeploymentSpec: llmEngineDeploymentSpec,
		llmEngine:            llmEngine,
		modelSpec:            modelSpec,
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
