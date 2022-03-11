package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="EXPIRES",type="string",JSONPath=".spec.expires"
// +kubebuilder:printcolumn:name="ALLOCATION STATUS",type="string",JSONPath=".status.allocationStatus"
// +kubebuilder:printcolumn:name="OPERATIONAL STATUS",type="string",JSONPath=".status.operationalStatus"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:singular=sliver,path=slivers,scope=Namespaced
type Sliver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SliverSpec   `json:"spec,omitempty"`
	Status            SliverStatus `json:"status,omitempty"`
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
}

type SliverStatus struct {
	AllocationStatus  string `json:"allocationStatus"`
	OperationalStatus string `json:"operationalStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SliverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Sliver `json:"items"`
}
