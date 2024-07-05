package cloud

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/services"
	"github.com/vexxhost/openstack-service-operator/api"
)

type FindServiceOpts struct {
	ID   *string
	Name string
	Type string
}

func FindService(ctx context.Context, client *gophercloud.ServiceClient, opts *FindServiceOpts) (*services.Service, error) {
	if opts.ID != nil {
		return services.Get(ctx, client, *opts.ID).Extract()
	}

	pages, err := services.List(client, services.ListOpts{
		Name:        opts.Name,
		ServiceType: opts.Type,
	}).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	matchedServices, err := services.ExtractServices(pages)
	if err != nil {
		return nil, err
	}

	if len(matchedServices) == 0 {
		return nil, api.ErrorNotFound
	} else if len(matchedServices) > 1 {
		return nil, api.ErrorMultipleFound
	}

	return &matchedServices[0], nil
}
