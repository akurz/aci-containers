package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json
// tags for the fields to be serialized.

// SnatPolicySpec defines the desired state of SnatPolicy
// +k8s:openapi-gen=true
type SnatPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code
	// after modifying this file
	// Add custom validation using kubebuilder tags:
	// https://book.kubebuilder.io/beyond_basics/generating_crd.html
	SnatIp    []string    `json:"snatIp"`
	Selector  PodSelector `json:"selector,omitempty"`
	PortRange []PortRange `json:"portRange"`
	// +kubebuilder:validation:Enum=tcp,udp,icmp
	Protocols []string `json:"protocols,omitempty"`
}

// SnatPolicyStatus defines the observed state of SnatPolicy
// +k8s:openapi-gen=true
type SnatPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	SnatPortsAllocated map[string][]NodePortRange `json:"snat-allocation,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnatPolicy is the Schema for the snatpolicies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SnatPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnatPolicySpec   `json:"spec,omitempty"`
	Status SnatPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnatPolicyList contains a list of SnatPolicy
type SnatPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnatPolicy `json:"items"`
}

type PortRange struct {
	Start int `json:"start,omitempty"`
	End   int `json:"end,omitempty"`
}

type PodSelector struct {
	Labels     []Label `json:"labels,omitempty"`
	Deployment string  `json:"deployment,omitempty"`
	Namespace  string  `json:"namespace,omitempty"`
}

type Label struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type NodePortRange struct {
	NodeName  string    `json:"nodename,omitempty"`
	PortRange PortRange `json:"portrange,omitempty"`
}

//func init() {
//	(*runtime.SchemeBuilder).Register(&SnatPolicy{}, &SnatPolicyList{})
//}