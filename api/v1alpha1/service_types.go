// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// +kubebuilder:validation:Required
	Description string `json:"description"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	Ready     bool    `json:"ready"`
	ServiceID *string `json:"serviceID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

func (s *Service) UpdateOpenStackResource(ctx context.Context, client *gophercloud.ServiceClient) (*services.Service, error) {
	return services.Update(ctx, client, *s.Status.ServiceID, services.UpdateOpts{
		Type: s.Spec.Type,
		Extra: map[string]interface{}{
			"name":        s.Spec.Name,
			"description": s.Spec.Description,
		},
	}).Extract()
}

//+kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
