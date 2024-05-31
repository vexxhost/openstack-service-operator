# SPDX-License-Identifier: Apache-2.0

import kopf
import openstack


def adopt_openstack_resource(patch: kopf.Patch, resource_id: str):
    patch.status["resourceId"] = resource_id


def get_openstack_resource_id(
    status: dict,
):
    return status.get("resourceId")


def get_openstack_client():
    return openstack.connect(cloud="envvars")
