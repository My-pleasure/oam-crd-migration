package converter

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Converter is the plugin that convert v1alpha1 OAM types to v1alpha2 types
type Converter interface {

	// ConvertAppConfig converts v1alpha1 AppConfig to v1alpha2
	ConvertAppConfig(Object *unstructured.Unstructured)
}

type converts struct {
}

func (c *converts) ConvertAppConfig(Object *unstructured.Unstructured) {

}
