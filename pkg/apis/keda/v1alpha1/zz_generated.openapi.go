// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaController":       schema_pkg_apis_keda_v1alpha1_KedaController(ref),
		"github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerSpec":   schema_pkg_apis_keda_v1alpha1_KedaControllerSpec(ref),
		"github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerStatus": schema_pkg_apis_keda_v1alpha1_KedaControllerStatus(ref),
	}
}

func schema_pkg_apis_keda_v1alpha1_KedaController(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "KedaController is the Schema for the kedacontrollers API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerSpec", "github.com/kedacore/keda-olm-operator/pkg/apis/keda/v1alpha1.KedaControllerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_keda_v1alpha1_KedaControllerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "KedaControllerSpec defines the desired state of KedaController",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"logLevel": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"logLevelMetrics": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"watchNamespace": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_keda_v1alpha1_KedaControllerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "KedaControllerStatus defines the observed state of KedaController",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"phase": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"reason": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"phase"},
			},
		},
	}
}