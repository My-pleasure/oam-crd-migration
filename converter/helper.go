package converter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Converter is the plugin that convert v1alpha1 OAM types to v1alpha2 types
type Converter interface {

	// ConvertExampleCRD converts v1alpha1 Example to v1alpha2
	ConvertExampleCRD(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status)

	// ConvertAppConfig converts v1alpha1 AppConfig to v1alpha2
	ConvertAppConfig(Object *unstructured.Unstructured) (*unstructured.Unstructured, metav1.Status)
}

type OAMConverts struct {
}
