/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CachingPolicySpec defines the desired state of CachingPolicy
type CachingPolicySpec struct {
	PolicySpec    `json:"",inline`
	CachingPolicy string `json:"cachingPolicy"`
}

// CachingPolicyStatus defines the observed state of CachingPolicy
type CachingPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClustersDone []string `json:"clustersDone,omitempty"`
	// +kubebuilder:validation:Enum:=Success;Fail
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CachingPolicy is the Schema for the cachingpolicies API
type CachingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CachingPolicySpec   `json:"spec,omitempty"`
	Status CachingPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CachingPolicyList contains a list of CachingPolicy
type CachingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CachingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CachingPolicy{}, &CachingPolicyList{})
}
