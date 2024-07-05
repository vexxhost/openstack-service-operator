package cloud

import (
	"context"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/endpoints"
	"github.com/vexxhost/openstack-service-operator/api"
)

type FindEndpointOpts struct {
	ID        *string
	Region    string
	ServiceID *string
	Interface gophercloud.Availability
}

func FindEndpoint(ctx context.Context, client *gophercloud.ServiceClient, opts *FindEndpointOpts) (*endpoints.Endpoint, error) {
	pages, err := endpoints.List(client, endpoints.ListOpts{
		RegionID:     opts.Region,
		ServiceID:    *opts.ServiceID,
		Availability: opts.Interface,
	}).AllPages(ctx)
	if err != nil {
		return nil, err
	}

	matchedEndpoints, err := endpoints.ExtractEndpoints(pages)
	if err != nil {
		return nil, err
	}

	if len(matchedEndpoints) == 0 {
		return nil, api.ErrorNotFound
	}

	if opts.ID != nil {
		for _, endpoint := range matchedEndpoints {
			if endpoint.ID == *opts.ID {
				return &endpoint, nil
			}
		}

		return nil, api.ErrorNotFound
	}

	if len(matchedEndpoints) > 1 {
		return nil, api.ErrorMultipleFound
	}

	return &matchedEndpoints[0], nil
}
