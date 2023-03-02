/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StoredFunctionSpec defines the desired state of StoredFunction
type StoredFunctionSpec struct {
	// +kubebuilder:validation:MinItems:=1
	ClusterUris []string `json:"clusterUris"`
	DB          string   `json:"db"`
	// Name is the name of the function
	Name string `json:"name"`
	// +kubebuilder:validation:Optional
	// DocString is the function documentation, optional
	DocString string `json:"docString,omitempty"`
	// +kubebuilder:validation:Optional
	// Folder is the function folder, optional
	Folder string `json:"folder,omitempty"`
	// +kubebuilder:validation:Optional
	// Parameters is the function parameters, optional
	Parameters string `json:"parameters,omitempty"`
	// Body is the function body
	Body string `json:"body"`
}

// StoredFunctionStatus defines the observed state of StoredFunction
type StoredFunctionStatus struct {
	ClustersDone []string `json:"clustersDone,omitempty"`
	// +kubebuilder:validation:Enum:=Success;Fail
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StoredFunction is the Schema for the storedfunctions API
type StoredFunction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoredFunctionSpec   `json:"spec,omitempty"`
	Status StoredFunctionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoredFunctionList contains a list of StoredFunction
type StoredFunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoredFunction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoredFunction{}, &StoredFunctionList{})
}
