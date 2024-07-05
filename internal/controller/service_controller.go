// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/services"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vexxhost/openstack-service-operator/api"
	identityv1alpha1 "github.com/vexxhost/openstack-service-operator/api/v1alpha1"
	"github.com/vexxhost/openstack-service-operator/internal/cloud"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	identityClient *gophercloud.ServiceClient
}

//+kubebuilder:rbac:groups=identity.openstack.org,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=identity.openstack.org,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=identity.openstack.org,resources=services/finalizers,verbs=update

func (r *ServiceReconciler) serviceNeedsUpdate(object *identityv1alpha1.Service, service *services.Service) bool {
	return object.Spec.Name != service.Extra["name"].(string) ||
		object.Spec.Type != service.Type ||
		object.Spec.Description != service.Extra["description"].(string)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	kubeService := &identityv1alpha1.Service{}
	if err := r.Get(ctx, req.NamespacedName, kubeService); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	osService, err := cloud.FindService(ctx, r.identityClient, &cloud.FindServiceOpts{
		ID:   kubeService.Status.ServiceID,
		Name: kubeService.Spec.Name,
		Type: kubeService.Spec.Type,
	})

	if errors.Is(err, api.ErrorNotFound) {
		osService, err = services.Create(ctx, r.identityClient, services.CreateOpts{
			Type: kubeService.Spec.Type,
			Extra: map[string]interface{}{
				"name":        kubeService.Spec.Name,
				"description": kubeService.Spec.Description,
			},
		}).Extract()

		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if kubeService.Spec.Name != osService.Extra["name"].(string) || kubeService.Spec.Type != osService.Type || kubeService.Spec.Description != osService.Extra["description"].(string) {
		_, err = services.Update(ctx, r.identityClient, *kubeService.Status.ServiceID, services.UpdateOpts{
			Type: kubeService.Spec.Type,
			Extra: map[string]interface{}{
				"name":        kubeService.Spec.Name,
				"description": kubeService.Spec.Description,
			},
		}).Extract()

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	kubeService.Status.Ready = true
	kubeService.Status.ServiceID = &osService.ID

	if err := r.Status().Update(ctx, kubeService); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&identityv1alpha1.Service{}).
		Complete(r)
}
