package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Fallback struct {
	// Replicas is the number of replicas to fallback to. The is manifested as
	// patching the HorizontalPodAutoscaler.spec.minReplicas.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Default=1
	Replicas int32 `json:"replicas,omitempty"`

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
}

// HorizontalPodAutoscalerXStatus defines the observed state of HorizontalPodAutoscalerX.
type HorizontalPodAutoscalerXStatus struct{}

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
