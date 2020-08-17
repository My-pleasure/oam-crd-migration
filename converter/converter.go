package converter

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

func ConvertAppConfig(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status) {
	klog.V(2).Info("converting crd")

	convertedObject := Object.DeepCopy()
	fromVersion := Object.GetAPIVersion()

	if toVersion == fromVersion {
		return nil, statusErrorWithMessage("conversion from a version to itself should not call the webhook: %s", toVersion)
	}

	converterPlugin := Plugin{}

	switch Object.GetAPIVersion() {
	case "core.oam.dev/v1alpha1":
		switch toVersion {
		case "core.oam.dev/v1alpha2":
			components, _, err := unstructured.NestedSlice(convertedObject.Object, "spec", "components")
			if err != nil {
				return nil, statusErrorWithMessage("get spec.components error: ", err)
			}

			v1alpha2Components := make([]interface{}, 0)
			for _, comp := range components {
				c, _, err := converterPlugin.ConvertComponent(comp)
				if err != nil {
					return nil, statusErrorWithMessage("convert components error: ", err)
				}
				// TODO: apply the v1alpha2 component

				traits, _, _ := unstructured.NestedSlice(c, "traits")

				v1alpha2Traits := make([]interface{}, 0)
				for _, tr := range traits {
					v1alpha2Trait, err := converterPlugin.ConvertTrait(tr)
					if err != nil {
						return nil, statusErrorWithMessage("convert trait error: ", err)
					}

					tempTrait := make(map[string]interface{}, 0)
					_ = unstructured.SetNestedField(tempTrait, v1alpha2Trait, "trait")
					v1alpha2Traits = append(v1alpha2Traits, tempTrait)
				}

				unstructured.RemoveNestedField(c, "traits")
				err = unstructured.SetNestedSlice(c, v1alpha2Traits, "traits")
				if err != nil {
					return nil, statusErrorWithMessage("set component.traits error: ", err)
				}

				v1alpha2Components = append(v1alpha2Components, c)
			}

			unstructured.RemoveNestedField(convertedObject.Object, "spec", "components")
			err = unstructured.SetNestedSlice(convertedObject.Object, v1alpha2Components, "spec", "components")
			if err != nil {
				klog.Info("set spec.components err: ", err)
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	case "core.oam.dev/v1alpha2":
		switch toVersion {
		case "core.oam.dev/v1alpha1":
			//
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version %q", fromVersion)
	}

	return convertedObject, statusSucceed()
}
