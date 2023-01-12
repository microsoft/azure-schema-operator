// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VersionedDeplymentSpec defines the desired state of VersionedDeplyment
type VersionedDeplymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VersionedDeplyment. Edit versioneddeplyment_types.go to remove/update
	Revision       int32          `json:"revision"`
	ConfigMapName  NamespacedName `json:"configMapName"`
	ApplyTo        TargetFilter   `json:"applyTo"`
	Type           DBTypeEnum     `json:"type"`
	FailIfDataLoss bool           `json:"failIfDataLoss"`
}

// VersionedDeplymentStatus defines the observed state of VersionedDeplyment
type VersionedDeplymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Executers    []NamespacedName `json:"executers"`
	Executed     bool             `json:"executed"`
	Failed       int32            `json:"failed"`
	Running      int32            `json:"running"`
	Succeeded    int32            `json:"succeeded"`
	CompletedPCT int              `json:"completedPct,omitempty"`
	// Conditions is an array of conditions.
	// Known .status.conditions.type are: "Execution"
	//+patchMergeKey=type
	//+patchStrategy=merge
	//+listType=map
	//+listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// VersionedDeplyment is an immutable object that represents a deployment of a specific revision of a schema deployment
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CompletedPCT",type="string",JSONPath=".status.completedPct"
type VersionedDeplyment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VersionedDeplymentSpec   `json:"spec,omitempty"`
	Status VersionedDeplymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VersionedDeplymentList contains a list of VersionedDeplyment
type VersionedDeplymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VersionedDeplyment `json:"items"`
}

// IsExecuted checks the executed status
func (t *VersionedDeplyment) IsExecuted() bool {
	return t.Status.Executed
}

// IsRunning checks the running status
func (t *VersionedDeplyment) IsRunning() bool {
	return t.Status.Running > 0
}

// IsFailed checks the failed status
func (t *VersionedDeplyment) IsFailed() bool {
	return t.Status.Failed > 0
}

func init() {
	SchemeBuilder.Register(&VersionedDeplyment{}, &VersionedDeplymentList{})
}
