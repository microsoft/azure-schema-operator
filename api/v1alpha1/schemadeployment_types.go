// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FailurePolicyEnum Enum for the different failure policies
// +kubebuilder:validation:Enum=abort;ignore;rollback
type FailurePolicyEnum string

const (
	// FailurePolicyAbort abort policy stops the execution in failed state
	FailurePolicyAbort FailurePolicyEnum = "abort"
	// FailurePolicyIgnore ignore policy ignores the errors and finish the execution successfully.
	FailurePolicyIgnore FailurePolicyEnum = "ignore"
	// FailurePolicyRollback rollback policy rolls back the configuration to a previous one.
	FailurePolicyRollback FailurePolicyEnum = "rollback"
)

// DBTypeEnum Enum for the supported DB types
type DBTypeEnum string

const (
	// DBTypeSQLServer SQL Server enum entry
	DBTypeSQLServer DBTypeEnum = "sqlServer"
	// DBTypeKusto Kusto enum entry
	DBTypeKusto DBTypeEnum = "kusto"
	// DBTypeEventhub eventhub schema registry enum entry
	DBTypeEventhub DBTypeEnum = "eventhub"
	// ConditionExecuted execution condition status
	ConditionExecution string = "Execution"
)

// TargetFilter contains target filter configuration
type TargetFilter struct {
	// +kubebuilder:validation:MinItems:=1
	ClusterUris []string `json:"clusterUris"`
	Schema      string   `json:"schema,omitempty"`
	DB          string   `json:"db"`
	// +kubebuilder:validation:Optional
	Webhook string `json:"webhook,omitempty"`
	// +kubebuilder:validation:Optional
	Label string `json:"label,omitempty"`
	// +kubebuilder:validation:Optional
	DBS    []string `json:"dbs,omitempty"`
	Create bool     `json:"create,omitempty"`
	Regexp bool     `json:"regexp,omitempty"`
}

// SchemaDeploymentSpec defines the desired state of SchemaDeployment
type SchemaDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ApplyTo TargetFilter   `json:"applyTo"`
	Type    DBTypeEnum     `json:"type"`
	Source  NamespacedName `json:"source,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=rollback
	FailurePolicy FailurePolicyEnum `json:"failurePolicy"`
	// +kubebuilder:default:=true
	FailIfDataLoss bool `json:"failIfDataLoss"`
}

// SchemaDeploymentStatus defines the observed state of SchemaDeployment
type SchemaDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Executed               bool             `json:"executed"`
	DesiredNumberScheduled int32            `json:"desiredNumberScheduled"`
	CurrentConfigMap       NamespacedName   `json:"currentConfigMap"`
	LastConfigMap          string           `json:"lastConfigMap"`
	CurrentRevision        int32            `json:"currentRevision"`
	LastSuccessfulRevision int32            `json:"lastSuccessfulRevision"`
	CurrentVerDeployment   NamespacedName   `json:"currentVerDeployment"`
	OldVerDeployment       []NamespacedName `json:"oldVerDeployment,omitempty"`
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
//+kubebuilder:printcolumn:name="Executed",type="string",JSONPath=".status.conditions[?(@.type=='Executed')].status"
// SchemaDeployment is the Schema for the templates API
type SchemaDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SchemaDeploymentSpec   `json:"spec,omitempty"`
	Status SchemaDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SchemaDeploymentList contains a list of SchemaDeployment
type SchemaDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SchemaDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SchemaDeployment{}, &SchemaDeploymentList{})
}

// IsExecuted checks if the schema deployment object was executed.
func (t *SchemaDeployment) IsExecuted() bool {
	return t.Status.Executed
}
