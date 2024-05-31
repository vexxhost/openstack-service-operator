# SPDX-License-Identifier: Apache-2.0

import kopf

from openstack_service_operator.operators import common

# positional arguments:
#   <service>             Service to be associated with new endpoint (name or
#                         ID)
#   <interface>           New endpoint interface type (admin, public or
#                         internal)
#   <url>                 New endpoint URL


@kopf.on.resume("identity.openstack.vexxhost.com", "v1alpha1", "endpoint")
@kopf.on.create("identity.openstack.vexxhost.com", "v1alpha1", "endpoint")
def create_fn(logger, spec, status, patch, **_):
    endpoint_id = common.get_openstack_resource_id(status)
    if endpoint_id:
        logger.info("identity.Endpoint ID already set, skipping")
        return

    conn = common.get_openstack_client()
    endpoints = list(conn.identity.endpoints(
        service=spec["service"], interface=spec["interface"], url=spec["url"]
    ))
    if not endpoints:
        endpoint = conn.identity.create_endpoint(
            service=spec["service"],
            interface=spec["interface"],
            url=spec["url"],
        )
        logger.info("identity.Endpoint created in OpenStack")
    else:
        if len(endpoints) > 1:
            raise kopf.PermanentError(
                "Multiple endpoints found for the same service and interface"
            )

        endpoint = endpoints[0]
        logger.info("identity.Endpoint already exists in OpenStack, adopting")

    common.adopt_openstack_resource(patch=patch, resource_id=endpoint.id)


@kopf.on.update("identity.openstack.vexxhost.com", "v1alpha1", "endpoint")
def update_fn(logger, spec, status, **_):
    endpoint_id = common.get_openstack_resource_id(status)
    if endpoint_id is None:
        raise kopf.TemporaryError("identity.Endpoint ID is not set, unable to update")

    conn = common.get_openstack_client()
    conn.identity.update_endpoint(
        endpoint_id,
        service=spec["service"],
        interface=spec["interface"],
        url=spec["url"],
    )
    logger.info("identity.Endpoint updated in OpenStack")


@kopf.on.delete("identity.openstack.vexxhost.com", "v1alpha1", "endpoint")
def delete_fn(logger, status, **_):
    endpoint_id = common.get_openstack_resource_id(status)
    if endpoint_id is None:
        logger.info("identity.Endpoint ID is not set, nothing to delete")
        return

    conn = common.get_openstack_client()
    conn.identity.delete_endpoint(endpoint_id, ignore_missing=True)
    logger.info("identity.Endpoint deleted")
