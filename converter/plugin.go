package converter

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/crossplane/oam-kubernetes-runtime/apis/core/v1alpha2"
)

type v1alpha1Component interface{}
type v1alpha2Component map[string]interface{}
type v1alpha2ComponentCR v1alpha2.Component
type v1alpha1Trait interface{}
type v1alpha2Trait map[string]interface{}

// Converter is the plugin that convert v1alpha1 OAM types to v1alpha2 types
type Converter interface {
	// ConvertComponent converts spec.components from v1alpha1 types to v1alpha2 types
	// in ApplicationConfigurations and return a v1alpha2 component CR.
	ConvertComponent(v1alpha1Component) (v1alpha2Component, v1alpha2ComponentCR, error)

	// ConvertTrait converts spec.components[*].traits from v1alpha1 types to v1alpha2 types
	// in ApplicationConfigurations.
	ConvertTrait(v1alpha1Trait) (v1alpha2Trait, error)
}

type Plugin struct {
}

func (p *Plugin) ConvertComponent(comp v1alpha1Component) (v1alpha2Component, v1alpha2ComponentCR, error) {
	c, _ := comp.(map[string]interface{})

	var v1alpha2Comp v1alpha2ComponentCR

	name, _, _ := unstructured.NestedString(c, "componentName")
	v1alpha2Comp.Name = name
	//parameterValues, _, _ := unstructured.NestedSlice(c, "parameterValues")
	//for _, value := range parameterValues {
	//	v, _ := value.(map[string]interface{})
	//}

	unstructured.RemoveNestedField(c, "parameterValues")
	unstructured.RemoveNestedField(c, "instanceName")

	return c, v1alpha2Comp, nil
}

func (p *Plugin) ConvertTrait(tr v1alpha1Trait) (v1alpha2Trait, error) {
	t, _ := tr.(map[string]interface{})
	v1alpha2Trait := make(map[string]interface{}, 0)

	_ = unstructured.SetNestedField(v1alpha2Trait, "core.oam.dev/v1alpha2", "apiVersion")
	_ = unstructured.SetNestedField(v1alpha2Trait, "RolloutTrait", "kind")

	v1alpha2Metadata := make(map[string]interface{}, 0)
	name, _, err := unstructured.NestedString(t, "name")
	if err != nil {
		return nil, err
	}
	_ = unstructured.SetNestedField(v1alpha2Metadata, name, "name")
	_ = unstructured.SetNestedField(v1alpha2Trait, v1alpha2Metadata, "metadata")

	v1alpha2Spec := make(map[string]interface{}, 0)
	properties, _, err := unstructured.NestedSlice(t, "properties")
	if err != nil {
		return nil, err
	}
	for _, pro := range properties {
		p, _ := pro.(map[string]interface{})

		name, ok, err := unstructured.NestedString(p, "name")
		if err != nil {
			return nil, err
		}
		if ok {
			value, _, _ := unstructured.NestedString(p, "value")
			_ = unstructured.SetNestedField(v1alpha2Spec, value, name)
		}
	}
	_ = unstructured.SetNestedField(v1alpha2Trait, v1alpha2Spec, "spec")

	return v1alpha2Trait, nil
}
