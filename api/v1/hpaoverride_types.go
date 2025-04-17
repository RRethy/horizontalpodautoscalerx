package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPAOverrideSpec defines the desired state of HPAOverride.
type HPAOverrideSpec struct {
	// MinReplicas is the minReplicas to override.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// Duration is the duration to apply this override.
	// +kubebuilder:validation:Required
	Duration metav1.Duration `json:"duration,omitempty"`

	// Time is the time to apply this override.
	// +kubebuilder:validation:Required
	Time metav1.Time `json:"time,omitempty"`

	// HPATargetName is the name of the HorizontalPodAutoscaler to override.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	HPATargetName string `json:"hpaTargetName,omitempty"`
}

// HPAOverrideStatus defines the observed state of HPAOverride.
type HPAOverrideStatus struct {
	// Active is the active status of the override.
	// +kubebuilder:validation:Optional
	Active bool `json:"active,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=all,shortName=hpao
// +kubebuilder:printcolumn:name="HPA",type=string,JSONPath=".spec.hpaTargetName",description="The name of the HorizontalPodAutoscaler to scale"
// +kubebuilder:printcolumn:name="MinReplicas",type=integer,JSONPath=".spec.minReplicas",description="The minReplicas to override"
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=".status.active",description="The active status of the override"

// HPAOverride is the Schema for the hpaoverrides API.
type HPAOverride struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HPAOverrideSpec   `json:"spec,omitempty"`
	Status HPAOverrideStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HPAOverrideList contains a list of HPAOverride.
type HPAOverrideList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HPAOverride `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HPAOverride{}, &HPAOverrideList{})
}
