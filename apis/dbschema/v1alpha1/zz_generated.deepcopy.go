//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterExecuter) DeepCopyInto(out *ClusterExecuter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterExecuter.
func (in *ClusterExecuter) DeepCopy() *ClusterExecuter {
	if in == nil {
		return nil
	}
	out := new(ClusterExecuter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterExecuter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterExecuterList) DeepCopyInto(out *ClusterExecuterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterExecuter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterExecuterList.
func (in *ClusterExecuterList) DeepCopy() *ClusterExecuterList {
	if in == nil {
		return nil
	}
	out := new(ClusterExecuterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterExecuterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterExecuterSpec) DeepCopyInto(out *ClusterExecuterSpec) {
	*out = *in
	in.ApplyTo.DeepCopyInto(&out.ApplyTo)
	out.ConfigMapName = in.ConfigMapName
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterExecuterSpec.
func (in *ClusterExecuterSpec) DeepCopy() *ClusterExecuterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterExecuterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterExecuterStatus) DeepCopyInto(out *ClusterExecuterStatus) {
	*out = *in
	in.Targets.DeepCopyInto(&out.Targets)
	in.DoneTargets.DeepCopyInto(&out.DoneTargets)
	in.Config.DeepCopyInto(&out.Config)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterExecuterStatus.
func (in *ClusterExecuterStatus) DeepCopy() *ClusterExecuterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterExecuterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterTargets) DeepCopyInto(out *ClusterTargets) {
	*out = *in
	if in.DBs != nil {
		in, out := &in.DBs, &out.DBs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Schemas != nil {
		in, out := &in.Schemas, &out.Schemas
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterTargets.
func (in *ClusterTargets) DeepCopy() *ClusterTargets {
	if in == nil {
		return nil
	}
	out := new(ClusterTargets)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecutionConfiguration) DeepCopyInto(out *ExecutionConfiguration) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecutionConfiguration.
func (in *ExecutionConfiguration) DeepCopy() *ExecutionConfiguration {
	if in == nil {
		return nil
	}
	out := new(ExecutionConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaDeployment) DeepCopyInto(out *SchemaDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaDeployment.
func (in *SchemaDeployment) DeepCopy() *SchemaDeployment {
	if in == nil {
		return nil
	}
	out := new(SchemaDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SchemaDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaDeploymentList) DeepCopyInto(out *SchemaDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SchemaDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaDeploymentList.
func (in *SchemaDeploymentList) DeepCopy() *SchemaDeploymentList {
	if in == nil {
		return nil
	}
	out := new(SchemaDeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SchemaDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaDeploymentSpec) DeepCopyInto(out *SchemaDeploymentSpec) {
	*out = *in
	in.ApplyTo.DeepCopyInto(&out.ApplyTo)
	out.Source = in.Source
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaDeploymentSpec.
func (in *SchemaDeploymentSpec) DeepCopy() *SchemaDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(SchemaDeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaDeploymentStatus) DeepCopyInto(out *SchemaDeploymentStatus) {
	*out = *in
	out.CurrentConfigMap = in.CurrentConfigMap
	out.CurrentVerDeployment = in.CurrentVerDeployment
	if in.OldVerDeployment != nil {
		in, out := &in.OldVerDeployment, &out.OldVerDeployment
		*out = make([]NamespacedName, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaDeploymentStatus.
func (in *SchemaDeploymentStatus) DeepCopy() *SchemaDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(SchemaDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetFilter) DeepCopyInto(out *TargetFilter) {
	*out = *in
	if in.ClusterUris != nil {
		in, out := &in.ClusterUris, &out.ClusterUris
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DBS != nil {
		in, out := &in.DBS, &out.DBS
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetFilter.
func (in *TargetFilter) DeepCopy() *TargetFilter {
	if in == nil {
		return nil
	}
	out := new(TargetFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionedDeplyment) DeepCopyInto(out *VersionedDeplyment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionedDeplyment.
func (in *VersionedDeplyment) DeepCopy() *VersionedDeplyment {
	if in == nil {
		return nil
	}
	out := new(VersionedDeplyment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VersionedDeplyment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionedDeplymentList) DeepCopyInto(out *VersionedDeplymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VersionedDeplyment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionedDeplymentList.
func (in *VersionedDeplymentList) DeepCopy() *VersionedDeplymentList {
	if in == nil {
		return nil
	}
	out := new(VersionedDeplymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VersionedDeplymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionedDeplymentSpec) DeepCopyInto(out *VersionedDeplymentSpec) {
	*out = *in
	out.ConfigMapName = in.ConfigMapName
	in.ApplyTo.DeepCopyInto(&out.ApplyTo)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionedDeplymentSpec.
func (in *VersionedDeplymentSpec) DeepCopy() *VersionedDeplymentSpec {
	if in == nil {
		return nil
	}
	out := new(VersionedDeplymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionedDeplymentStatus) DeepCopyInto(out *VersionedDeplymentStatus) {
	*out = *in
	if in.Executers != nil {
		in, out := &in.Executers, &out.Executers
		*out = make([]NamespacedName, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionedDeplymentStatus.
func (in *VersionedDeplymentStatus) DeepCopy() *VersionedDeplymentStatus {
	if in == nil {
		return nil
	}
	out := new(VersionedDeplymentStatus)
	in.DeepCopyInto(out)
	return out
}
