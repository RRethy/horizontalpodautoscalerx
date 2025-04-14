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

type HorizontalPodAutoscalerXConditionType string

const (
	// ConditionReady indicates that the HorizontalPodAutoscalerX is ready.
	ConditionReady HorizontalPodAutoscalerXConditionType = "Ready"
	// ConditionFallback indicates that the HorizontalPodAutoscalerX is in fallback mode.
	ConditionFallback HorizontalPodAutoscalerXConditionType = "FallbackTriggered"
)

// Condition represents the condition of the HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXCondition struct {
	// Type is the type of the condition.
	// +kubebuilder:validation:Required
	Type HorizontalPodAutoscalerXConditionType `json:"type,omitempty"`

	// Status is the status of the condition.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status corev1.ConditionStatus `json:"status,omitempty"`

	// ObservedGeneration is the generation of the HorizontalPodAutoscalerX
	// when the condition was last observed.
	// +kubebuilder:validation:Required
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// HPAObservedGeneration is the generation of the HorizontalPodAutoscaler
	// when the condition was last observed.
	// +kubebuilder:validation:Optional
	HPAObservedGeneration *int64 `json:"hpaObservedGeneration,omitempty"`

	// LastTransitionTime is the last time the condition transitioned from one
	// status to another.
	// +kubebuilder:validation:Optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief reason for the condition's last transition.
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about the last
	// transition.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

// HorizontalPodAutoscalerXStatus defines the observed state of HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXStatus struct {
	// Conditions is a list of conditions that apply to the HorizontalPodAutoscalerX.
	// +kubebuilder:validation:Optional
	Conditions []HorizontalPodAutoscalerXCondition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the HorizontalPodAutoscalerX
	// when it was last observed.
	// +kubebuilder:validation:Optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// HPAObservedGeneration is the generation of the HorizontalPodAutoscaler
	// when the HorizontalPodAutoscalerX was last observed.
	// +kubebuilder:validation:Optional
	HPAObservedGeneration *int64 `json:"hpaObservedGeneration,omitempty"`
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
