package converter

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

func (c *OAMConverts) ConvertExampleCRD(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status) {
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

func (c *OAMConverts) ConvertAppConfigCRD(Object *unstructured.Unstructured, toVersion string) (*unstructured.Unstructured, metav1.Status) {
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
			// There is the incomplete conversion logic. So the v1alpha2 object will lose some field values.
			delete(convertedObject.Object, "spec.components[0].instanceName")
		default:
			return nil, statusErrorWithMessage("unexpected conversion version %q", toVersion)
		}
	default:
		return nil, statusErrorWithMessage("unexpected conversion version %q", fromVersion)
	}

	return convertedObject, statusSucceed()
}
