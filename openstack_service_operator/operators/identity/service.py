# SPDX-License-Identifier: Apache-2.0

import kopf

from openstack_service_operator.operators import common


@kopf.on.resume("identity.openstack.vexxhost.com", "v1alpha1", "service")
@kopf.on.create("identity.openstack.vexxhost.com", "v1alpha1", "service")
def create_fn(logger, spec, status, patch, **_):
    service_id = common.get_openstack_resource_id(status)
    if service_id:
        logger.info("identity.Service ID already set, skipping")
        return

    conn = common.get_openstack_client()
    service = conn.identity.find_service(spec["name"])
    if not service:
        service = conn.identity.create_service(
            name=spec["name"], type=spec["type"], description=spec["description"]
        )
        logger.info("identity.Service created in OpenStack")
    else:
        logger.info("identity.Service already exists in OpenStack, adopting")

    common.adopt_openstack_resource(patch=patch, resource_id=service.id)


@kopf.on.update("identity.openstack.vexxhost.com", "v1alpha1", "service")
def update_fn(logger, spec, status, **_):
    service_id = common.get_openstack_resource_id(status)
    if service_id is None:
        raise kopf.TemporaryError("identity.Service ID is not set, unable to update")

    conn = common.get_openstack_client()
    conn.identity.update_service(
        service_id,
        name=spec["name"],
        type=spec["type"],
        description=spec["description"],
    )
    logger.info("identity.Service updated in OpenStack")


@kopf.on.delete("identity.openstack.vexxhost.com", "v1alpha1", "service")
def delete_fn(logger, status, **_):
    service_id = common.get_openstack_resource_id(status)
    if service_id is None:
        logger.info("identity.Service ID is not set, nothing to delete")
        return

    conn = common.get_openstack_client()
    conn.identity.delete_service(service_id, ignore_missing=True)
    logger.info("identity.Service deleted")
