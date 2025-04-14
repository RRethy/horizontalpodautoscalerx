package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HPAOverrideSpec defines the desired state of HPAOverride.
type HPAOverrideSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of HPAOverride. Edit hpaoverride_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HPAOverrideStatus defines the observed state of HPAOverride.
type HPAOverrideStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
