package repository

import (
	"fmt"

	"github.com/lib/pq"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/isutton/orchid/pkg/orchid/orm"
)

// nestedSlice extract informed field path and converts as an PostgreSQL array.
func nestedSlice(obj map[string]interface{}, fieldPath []string) (interface{}, error) {
	slice, found, err := unstructured.NestedSlice(obj, fieldPath...)
	if !found {
		return nil, fmt.Errorf("unable to find data at '%#v'", fieldPath)
	}
	if err != nil {
		return nil, err
	}
	return pq.Array(slice), nil
}

// nestedBool extract informed field path as boolean.
func nestedBool(obj map[string]interface{}, fieldPath []string) (bool, error) {
	boolean, found, err := unstructured.NestedBool(obj, fieldPath...)
	if !found {
		return false, fmt.Errorf("unable to find data at '%#v'", fieldPath)
	}
	if err != nil {
		return false, err
	}
	return boolean, nil
}

// nestedString extracted informed filed path as string.
func nestedString(obj map[string]interface{}, fieldPath []string) (string, error) {
	str, found, err := unstructured.NestedString(obj, fieldPath...)
	if !found {
		return "", fmt.Errorf("unable to find names at '%#v'", fieldPath)
	}
	if err != nil {
		return "", err
	}
	return str, nil
}

// nestedInt64 extract informed field path as int64.
func nestedInt64(obj map[string]interface{}, fieldPath []string) (int64, error) {
	integer, found, err := unstructured.NestedInt64(obj, fieldPath...)
	if !found {
		return 0, fmt.Errorf("unable to find data at '%#v'", fieldPath)
	}
	if err != nil {
		return 0, err
	}
	return integer, nil
}

// nestedFloat64 extract informed field path as float64.
func nestedFloat64(obj map[string]interface{}, fieldPath []string) (float64, error) {
	number, found, err := unstructured.NestedFloat64(obj, fieldPath...)
	if !found {
		return 0, fmt.Errorf("unable to find data at '%#v'", fieldPath)
	}
	if err != nil {
		return 0, err
	}
	return number, nil
}

// extract informed field path based on informed original type. The original type is expected to be
// based on JSON-Schema types.
func extract(
	obj map[string]interface{},
	originalType string,
	fieldPath []string,
) (interface{}, error) {
	var data interface{}
	var err error

	switch originalType {
	case orm.JSTypeArray:
		data, err = nestedSlice(obj, fieldPath)
	case orm.JSTypeBoolean:
		data, err = nestedBool(obj, fieldPath)
	case orm.JSTypeString:
		data, err = nestedString(obj, fieldPath)
	case orm.JSTypeInteger:
		data, err = nestedInt64(obj, fieldPath)
	case orm.JSTypeNumber:
		data, err = nestedFloat64(obj, fieldPath)
	default:
		return nil, fmt.Errorf("unable to handle type '%s'", originalType)
	}
	if data == nil {
		return nil, fmt.Errorf("unable to extract data from field path '%#v'", fieldPath)
	}
	if err != nil {
		return nil, err
	}
	return data, err

}

// nestedMap extract informed field path as a Map.
func nestedMap(obj map[string]interface{}, fieldPath []string) (map[string]interface{}, error) {
	data, found, err := unstructured.NestedMap(obj, fieldPath...)
	if !found {
		return nil, fmt.Errorf("unable to find data at '%#v'", fieldPath)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// extractCRDOpenAPIV3Schema extract known field path to store OpenAPI schema in a CRD unstructured
// Object, and returns as an actual JSONSchemaProps.
func extractCRDOpenAPIV3Schema(obj map[string]interface{}) (*extv1beta1.JSONSchemaProps, error) {
	data, err := nestedMap(obj, []string{"spec", "validation", "openAPIV3Schema"})
	if err != nil {
		return nil, err
	}
	openAPIV3Schema := &extv1beta1.JSONSchemaProps{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(data, &openAPIV3Schema)
	if err != nil {
		return nil, err
	}
	return openAPIV3Schema, nil
}

// extractCRGVKFromCRD extract target CR GVK from a CRD object.
func extractCRGVKFromCRD(obj map[string]interface{}) (schema.GroupVersionKind, error) {
	gvk := schema.GroupVersionKind{}

	data, err := nestedMap(obj, []string{"spec"})
	if err != nil {
		return gvk, err
	}

	if group, found := data["group"]; !found {
		return gvk, fmt.Errorf("unable to find group")
	} else {
		gvk.Group = group.(string)
	}
	if version, found := data["version"]; !found {
		return gvk, fmt.Errorf("unable to find version")
	} else {
		gvk.Version = version.(string)
	}
	if kind, err := nestedString(obj, []string{"spec", "names", "kind"}); err != nil {
		return gvk, err
	} else {
		gvk.Kind = kind
	}

	return gvk, nil
}