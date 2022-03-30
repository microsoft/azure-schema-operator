// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespacedName is an object identifier
type NamespacedName struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ClusterTargets contains DB and Schema arrays to run the change on.
type ClusterTargets struct {
	DBs     []string `json:"dbs,omitempty"`
	Schemas []string `json:"schemas,omitempty"`
}

// ExecutionConfiguration contains the required configuration for execution
type ExecutionConfiguration struct {
	KQLFile      string            `json:"kqlfile,omitempty"`
	JobFile      string            `json:"jobfile,omitempty"`
	DacPac       string            `json:"dacpac,omitempty"`
	TemplateName string            `json:"templatename,omitempty"`
	Schema       string            `json:"schema,omitempty"`
	Group        string            `json:"group,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterExecuterSpec defines the desired state of ClusterExecuter
type ClusterExecuterSpec struct {
	ClusterUri     string         `json:"clusterUri,omitempty"`
	ApplyTo        TargetFilter   `json:"applyTo"`
	Type           DBTypeEnum     `json:"type"`
	ConfigMapName  NamespacedName `json:"configMapName"`
	FailIfDataLoss bool           `json:"failIfDataLoss"`
	Revision       int32          `json:"revision"`
}

// ClusterExecuterStatus defines the observed state of ClusterExecuter
type ClusterExecuterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Executed    bool                   `json:"executed"`
	Running     bool                   `json:"running"`
	Failed      bool                   `json:"failed"`
	Targets     ClusterTargets         `json:"targets"`
	DoneTargets ClusterTargets         `json:"done"`
	Config      ExecutionConfiguration `json:"config,omitempty"`
	NumFailures int                    `json:"numFailures,omitempty"`
	// Conditions is an array of conditions.
	// Known .status.conditions.type are: "Execution"
	//+patchMergeKey=type
	//+patchStrategy=merge
	//+listType=map
	//+listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.type"
//+kubebuilder:printcolumn:name="Executed",type="string",JSONPath=".status.conditions[?(@.type=='Execution')].status"
// ClusterExecuter is the Schema for the clusterexecuters API
type ClusterExecuter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterExecuterSpec   `json:"spec,omitempty"`
	Status ClusterExecuterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterExecuterList contains a list of ClusterExecuter
type ClusterExecuterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterExecuter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterExecuter{}, &ClusterExecuterList{})
}

// IsExecuted checks if the cluster executer already executed.
func (t *ClusterExecuter) IsExecuted() bool {
	return t.Status.Executed
}
