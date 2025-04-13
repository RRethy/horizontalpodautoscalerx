package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Fallback struct {
	// MinReplicas is the minReplicas to fallback to. The is manifested as
	// patching the HorizontalPodAutoscaler.spec.minReplicas.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Default=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// Duration is the minimum duration to observe a failing condition on the
	// HPA before triggering a fallback.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=0s
	Duration metav1.Duration `json:"duration,omitempty"`
}

// HorizontalPodAutoscalerXSpec defines the desired state of HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXSpec struct {
	// HPATargetName is the name of the HorizontalPodAutoscaler to scale.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	HPATargetName string `json:"hpaTargetName,omitempty"`

	// Fallback defines the fallback behavior.
	// +kubebuilder:validation:Optional
	Fallback *Fallback `json:"fallback,omitempty"`

	// MinReplicas is the minReplicas for the HPA.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Default=1
	MinReplicas int32 `json:"minReplicas,omitempty"`
}

// HorizontalPodAutoscalerXStatus defines the observed state of HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXStatus struct {
	// ScalingActiveConditionSince is the time when the scaling condition was last observed.
	// +kubebuilder:validation:Optional
	ScalingActiveConditionSince *metav1.Time `json:"scalingActiveConditionSince,omitempty"`

	// ScalingActiveCondition is the last observed scaling condition.
	// +kubebuilder:validation:Optional
	ScalingActiveCondition corev1.ConditionStatus `json:"scalingActiveCondition,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.minReplicas,statuspath=.status.minReplicas
// +kubebuilder:resource:categories=all,shortName=hpax
// +kubebuilder:printcolumn:name="HPA",type=string,JSONPath=".spec.hpaTargetName",description="The name of the HorizontalPodAutoscaler to scale"

// HorizontalPodAutoscalerX is the Schema for the horizontalpodautoscalerxes API.
type HorizontalPodAutoscalerX struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizontalPodAutoscalerXSpec   `json:"spec,omitempty"`
	Status HorizontalPodAutoscalerXStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HorizontalPodAutoscalerXList contains a list of HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HorizontalPodAutoscalerX `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HorizontalPodAutoscalerX{}, &HorizontalPodAutoscalerXList{})
}
