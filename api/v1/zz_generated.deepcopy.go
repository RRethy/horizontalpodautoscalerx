//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalPodAutoscalerX) DeepCopyInto(out *HorizontalPodAutoscalerX) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalPodAutoscalerX.
func (in *HorizontalPodAutoscalerX) DeepCopy() *HorizontalPodAutoscalerX {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerX)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalPodAutoscalerX) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalPodAutoscalerXList) DeepCopyInto(out *HorizontalPodAutoscalerXList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HorizontalPodAutoscalerX, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalPodAutoscalerXList.
func (in *HorizontalPodAutoscalerXList) DeepCopy() *HorizontalPodAutoscalerXList {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerXList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalPodAutoscalerXList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalPodAutoscalerXSpec) DeepCopyInto(out *HorizontalPodAutoscalerXSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalPodAutoscalerXSpec.
func (in *HorizontalPodAutoscalerXSpec) DeepCopy() *HorizontalPodAutoscalerXSpec {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerXSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalPodAutoscalerXStatus) DeepCopyInto(out *HorizontalPodAutoscalerXStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalPodAutoscalerXStatus.
func (in *HorizontalPodAutoscalerXStatus) DeepCopy() *HorizontalPodAutoscalerXStatus {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerXStatus)
	in.DeepCopyInto(out)
	return out
}
