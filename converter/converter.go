package converter

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	"github.com/crossplane/oam-kubernetes-runtime/apis/core/v1alpha2"
)

// Converter is the plugin that convert v1alpha1 OAM types to v1alpha2 types
type Converter interface {
	// ConvertComponent converts v1alpha1 Component in ApplicationConfiguration to
	// v1alpha2 Component and workload CR.
	ConvertComponent(interface{}) (map[string]interface{}, error)

	// ConvertTrait converts v1alpha1 Trait in ApplicationConfiguration to trait CR.
	ConvertTrait(interface{}) (map[string]interface{}, error)
}

type Plugin struct {
}

func (p *Plugin) ConvertComponent(comp interface{}) (map[string]interface{}, error) {
	c, _ := comp.(map[string]interface{})

	var v1alpha2Comp v1alpha2.Component

	name, _, _ := unstructured.NestedString(c, "componentName")
	v1alpha2Comp.Name = name

	//parameterValues, _, _ := unstructured.NestedSlice(c, "parameterValues")
	//for _, value := range parameterValues {
	//	v, _ := value.(map[string]interface{})
	//}

	unstructured.RemoveNestedField(c, "parameterValues")
	unstructured.RemoveNestedField(c, "instanceName")

	return c, nil
}

func (p *Plugin) ConvertTrait(tr interface{}) (map[string]interface{}, error) {
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

	// TODO: Use name and properties to create a trait CR whose version is v1alpha2

	return v1alpha2Trait, nil
}

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
				c, err := converterPlugin.ConvertComponent(comp)
				if err != nil {
					return nil, statusErrorWithMessage("convert components error: ", err)
				}

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
			// TODO
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version %q", fromVersion)
	}

	return convertedObject, statusSucceed()
}

// Just for example
func ConvertExampleCRD(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status) {
	klog.V(2).Info("converting crd")

	convertedObject := Object.DeepCopy()
	fromVersion := Object.GetAPIVersion()

	if toVersion == fromVersion {
		return nil, statusErrorWithMessage("conversion from a version to itself should not call the webhook: %s", toVersion)
	}

	switch Object.GetAPIVersion() {
	case "core.oam.dev/v1alpha1":
		switch toVersion {
		case "core.oam.dev/v1alpha2":
			hostPort, ok := convertedObject.Object["hostPort"]
			if ok {
				delete(convertedObject.Object, "hostPort")
				parts := strings.Split(hostPort.(string), ":")
				if len(parts) != 2 {
					return nil, statusErrorWithMessage("invalid hostPort value `%v`", hostPort)
				}
				convertedObject.Object["host"] = parts[0]
				convertedObject.Object["port"] = parts[1]
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	case "core.oam.dev/v1alpha2":
		switch toVersion {
		case "core.oam.dev/v1alpha1":
			host, hasHost := convertedObject.Object["host"]
			port, hasPort := convertedObject.Object["port"]
			if hasHost || hasPort {
				if !hasHost {
					host = ""
				}
				if !hasPort {
					port = ""
				}
				convertedObject.Object["hostPort"] = fmt.Sprintf("%s:%s", host, port)
				delete(convertedObject.Object, "host")
				delete(convertedObject.Object, "port")
			}
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version %q", fromVersion)
	}
	return convertedObject, statusSucceed()
}
