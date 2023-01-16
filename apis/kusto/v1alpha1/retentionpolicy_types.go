/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KustoRetentionPolicy defines a retention policy
type KustoRetentionPolicy struct {
	SoftDeletePeriod string `json:"softDeletePeriod"`
	// +kubebuilder:validation:Enum:=Disabled;Enabled
	Recoverability string `json:"recoverability"`
}

// RetentionPolicySpec defines the desired state of RetentionPolicy
type RetentionPolicySpec struct {
	// +kubebuilder:validation:MinItems:=1
	ClusterUris     []string             `json:"clusterUris"`
	DB              string               `json:"db"`
	Table           string               `json:"table"`
	RetentionPolicy KustoRetentionPolicy `json:"retentionPolicy"`
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
