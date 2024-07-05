// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/gophercloud/gophercloud/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EndpointSpec defines the desired state of Endpoint
type EndpointSpec struct {
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// +kubebuilder:validation:Required
	Interface gophercloud.Availability `json:"interface"`

	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// EndpointStatus defines the observed state of Endpoint
type EndpointStatus struct {
	// +kubebuilder:default=false
	Ready bool `json:"ready"`

	ServiceID  *string `json:"serviceID,omitempty"`
	EndpointID *string `json:"endpointID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Endpoint is the Schema for the endpoints API
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointSpec   `json:"spec,omitempty"`
	Status EndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
