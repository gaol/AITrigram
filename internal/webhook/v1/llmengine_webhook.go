/*
Copyright 2025.

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

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	aitrigramv1 "github.com/gaol/AITrigram/api/v1"
	"github.com/gaol/AITrigram/internal/controller"
)

// nolint:unused
// log is for logging in this package.
var llmenginelog = logf.Log.WithName("llmengine-resource")

// SetupLLMEngineWebhookWithManager registers the webhook for LLMEngine in the manager.
func SetupLLMEngineWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&aitrigramv1.LLMEngine{}).
		WithValidator(&LLMEngineCustomValidator{}).
		WithDefaulter(&LLMEngineCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-aitrigram-ihomeland-cn-v1-llmengine,mutating=true,failurePolicy=fail,sideEffects=None,groups=aitrigram.ihomeland.cn,resources=llmengines,verbs=create;update,versions=v1,name=mllmengine-v1.kb.io,admissionReviewVersions=v1

// LLMEngineCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind LLMEngine when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type LLMEngineCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &LLMEngineCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind LLMEngine.
func (d *LLMEngineCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	llmengine, ok := obj.(*aitrigramv1.LLMEngine)

	if !ok {
		return fmt.Errorf("expected an LLMEngine object but got %T", obj)
	}
	llmenginelog.Info("Defaulting for LLMEngine", "name", llmengine.GetName())

	defaultSpec := controller.DefaultLLMEngineSpec(&llmengine.Spec.EngineType)
	_spec, err := controller.MergeLLMSpecs(defaultSpec, &llmengine.Spec)
	if err != nil {
		return err
	}
	llmengine.Spec = *_spec
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-aitrigram-ihomeland-cn-v1-llmengine,mutating=false,failurePolicy=fail,sideEffects=None,groups=aitrigram.ihomeland.cn,resources=llmengines,verbs=create;update,versions=v1,name=vllmengine-v1.kb.io,admissionReviewVersions=v1

// LLMEngineCustomValidator struct is responsible for validating the LLMEngine resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type LLMEngineCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &LLMEngineCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type LLMEngine.
func (v *LLMEngineCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	llmengine, ok := obj.(*aitrigramv1.LLMEngine)
	if !ok {
		return nil, fmt.Errorf("expected a LLMEngine object but got %T", obj)
	}
	llmenginelog.Info("Validation for LLMEngine upon creation", "name", llmengine.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type LLMEngine.
func (v *LLMEngineCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	llmengine, ok := newObj.(*aitrigramv1.LLMEngine)
	if !ok {
		return nil, fmt.Errorf("expected a LLMEngine object for the newObj but got %T", newObj)
	}
	llmenginelog.Info("Validation for LLMEngine upon update", "name", llmengine.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type LLMEngine.
func (v *LLMEngineCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	llmengine, ok := obj.(*aitrigramv1.LLMEngine)
	if !ok {
		return nil, fmt.Errorf("expected a LLMEngine object but got %T", obj)
	}
	llmenginelog.Info("Validation for LLMEngine upon deletion", "name", llmengine.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
