/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1alpha1

import (
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicySpec defines the desired state of a Policy
type PolicySpec struct {
	// +kubebuilder:validation:MinItems:=1
	ClusterUris []string `json:"clusterUris"`
	DB          string   `json:"db"`
	Table       string   `json:"table"`
}

// RetentionPolicySpec defines the desired state of RetentionPolicy
type RetentionPolicySpec struct {
	PolicySpec      `json:"",inline`
	RetentionPolicy types.RetentionPolicy `json:"retentionPolicy"`
}

// RetentionPolicyStatus defines the observed state of RetentionPolicy
type RetentionPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClustersDone []string `json:"clustersDone,omitempty"`
	// +kubebuilder:validation:Enum:=Success;Fail
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RetentionPolicy is the Schema for the retentionpolicies API
type RetentionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RetentionPolicySpec   `json:"spec,omitempty"`
	Status RetentionPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RetentionPolicyList contains a list of RetentionPolicy
type RetentionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RetentionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RetentionPolicy{}, &RetentionPolicyList{})
}
