// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/endpoints"
	"github.com/vexxhost/openstack-service-operator/api"
	identityv1alpha1 "github.com/vexxhost/openstack-service-operator/api/v1alpha1"
	"github.com/vexxhost/openstack-service-operator/internal/cloud"
)

// EndpointReconciler reconciles a Endpoint object
type EndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	identityClient *gophercloud.ServiceClient
}

//+kubebuilder:rbac:groups=identity.openstack.org,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identity.openstack.org,resources=endpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identity.openstack.org,resources=endpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *EndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	kubeEndpoint := &identityv1alpha1.Endpoint{}
	if err := r.Get(ctx, req.NamespacedName, kubeEndpoint); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if kubeEndpoint.Status.ServiceID == nil {
		kubeService := &identityv1alpha1.Service{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: kubeEndpoint.Namespace, Name: kubeEndpoint.Spec.ServiceName}, kubeService); err != nil {
			return ctrl.Result{}, err
		}

		if !kubeService.Status.Ready {
			return ctrl.Result{}, errors.New("service is not ready")
		}

		kubeEndpoint.Status.ServiceID = kubeService.Status.ServiceID
		if err := r.Status().Update(ctx, kubeEndpoint); err != nil {
			return ctrl.Result{}, err
		}
	}

	osEndpoint, err := cloud.FindEndpoint(ctx, r.identityClient, &cloud.FindEndpointOpts{
		ID:        kubeEndpoint.Status.EndpointID,
		Region:    kubeEndpoint.Spec.Region,
		ServiceID: kubeEndpoint.Status.ServiceID,
		Interface: kubeEndpoint.Spec.Interface,
	})

	if errors.Is(err, api.ErrorNotFound) {
		osEndpoint, err = endpoints.Create(ctx, r.identityClient, endpoints.CreateOpts{
			Name:         kubeEndpoint.Name,
			Region:       kubeEndpoint.Spec.Region,
			ServiceID:    *kubeEndpoint.Status.ServiceID,
			Availability: kubeEndpoint.Spec.Interface,
			URL:          kubeEndpoint.Spec.URL,
		}).Extract()

		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if kubeEndpoint.Spec.URL != osEndpoint.URL {
		_, err := endpoints.Update(ctx, r.identityClient, osEndpoint.ID, endpoints.UpdateOpts{
			URL: kubeEndpoint.Spec.URL,
		}).Extract()

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	kubeEndpoint.Status.Ready = true
	kubeEndpoint.Status.EndpointID = &osEndpoint.ID

	if err := r.Status().Update(ctx, kubeEndpoint); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	provider, err := openstack.AuthenticatedClient(context.TODO(), opts)
	if err != nil {
		return err
	}

	r.identityClient, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&identityv1alpha1.Endpoint{}).
		Complete(r)
}
