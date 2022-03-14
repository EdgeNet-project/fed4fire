package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="SLICE URN",type="string",JSONPath=".spec.sliceUrn"
// +kubebuilder:printcolumn:name="EXPIRES",type="string",JSONPath=".spec.expires"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:singular=sliver,path=slivers,scope=Namespaced
type Sliver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SliverSpec `json:"spec,omitempty"`
}

type SliverSpec struct {
	// +kubebuilder:validation:Required
	URN string `json:"urn"`
	// +kubebuilder:validation:Required
	SliceURN string `json:"sliceUrn"`
	// +kubebuilder:validation:Required
	UserURN string `json:"userUrn"`
	// +kubebuilder:validation:Required
	Expires metav1.Time `json:"expires"`
	// +kubebuilder:validation:Required
	ClientID string `json:"clientId"`
	// +kubebuilder:validation:Required
	Image string `json:"image"`
	// +optional
	RequestedArch *string `json:"requestedArch"`
	// +optional
	RequestedNode *string `json:"requestedNode"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SliverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Sliver `json:"items"`
}
