//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Fallback) DeepCopyInto(out *Fallback) {
	*out = *in
	out.Duration = in.Duration
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fallback.
func (in *Fallback) DeepCopy() *Fallback {
	if in == nil {
		return nil
	}
	out := new(Fallback)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HPAOverride) DeepCopyInto(out *HPAOverride) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HPAOverride.
func (in *HPAOverride) DeepCopy() *HPAOverride {
	if in == nil {
		return nil
	}
	out := new(HPAOverride)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HPAOverride) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HPAOverrideList) DeepCopyInto(out *HPAOverrideList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HPAOverride, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HPAOverrideList.
func (in *HPAOverrideList) DeepCopy() *HPAOverrideList {
	if in == nil {
		return nil
	}
	out := new(HPAOverrideList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HPAOverrideList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HPAOverrideSpec) DeepCopyInto(out *HPAOverrideSpec) {
	*out = *in
	out.Duration = in.Duration
	in.Time.DeepCopyInto(&out.Time)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HPAOverrideSpec.
func (in *HPAOverrideSpec) DeepCopy() *HPAOverrideSpec {
	if in == nil {
		return nil
	}
	out := new(HPAOverrideSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HPAOverrideStatus) DeepCopyInto(out *HPAOverrideStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HPAOverrideStatus.
func (in *HPAOverrideStatus) DeepCopy() *HPAOverrideStatus {
	if in == nil {
		return nil
	}
	out := new(HPAOverrideStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalPodAutoscalerX) DeepCopyInto(out *HorizontalPodAutoscalerX) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
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
func (in *HorizontalPodAutoscalerXCondition) DeepCopyInto(out *HorizontalPodAutoscalerXCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalPodAutoscalerXCondition.
func (in *HorizontalPodAutoscalerXCondition) DeepCopy() *HorizontalPodAutoscalerXCondition {
	if in == nil {
		return nil
	}
	out := new(HorizontalPodAutoscalerXCondition)
	in.DeepCopyInto(out)
	return out
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
	if in.Fallback != nil {
		in, out := &in.Fallback, &out.Fallback
		*out = new(Fallback)
		**out = **in
	}
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
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]HorizontalPodAutoscalerXCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ObservedGeneration != nil {
		in, out := &in.ObservedGeneration, &out.ObservedGeneration
		*out = new(int64)
		**out = **in
	}
	if in.HPAObservedGeneration != nil {
		in, out := &in.HPAObservedGeneration, &out.HPAObservedGeneration
		*out = new(int64)
		**out = **in
	}
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
