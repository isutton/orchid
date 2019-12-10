package mocks

import (
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func JSONSchemaProps(
	jsonSchemaType string,
	format string,
	properties map[string]extv1beta1.JSONSchemaProps,
) extv1beta1.JSONSchemaProps {
	return extv1beta1.JSONSchemaProps{
		Type:       jsonSchemaType,
		Format:     format,
		Properties: properties,
	}
}

func JSONSchemaPropsComplex() map[string]extv1beta1.JSONSchemaProps {
	return map[string]extv1beta1.JSONSchemaProps{
		"simple": JSONSchemaProps("integer", "int32", nil),
		"complex": JSONSchemaProps("object", "", map[string]extv1beta1.JSONSchemaProps{
			"simple_nested": JSONSchemaProps("integer", "int32", nil),
			"complex_nested": JSONSchemaProps("object", "", map[string]extv1beta1.JSONSchemaProps{
				"attribute": JSONSchemaProps("string", "byte", nil),
			}),
		}),
	}
}