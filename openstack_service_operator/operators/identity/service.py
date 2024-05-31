# SPDX-License-Identifier: Apache-2.0

import logging

import kopf
import openstack


@kopf.on.resume("identity.openstack.vexxhost.com", "v1alpha1", "service")
@kopf.on.create("identity.openstack.vexxhost.com", "v1alpha1", "service")
def create_fn(body, **kwargs):
    conn = openstack.connect(cloud="envvars")
    conn.identity.create_service(
        name=body["metadata"]["name"],
        type=body["spec"]["type"],
        description=body["spec"]["description"],
    )


@kopf.on.delete("identity.openstack.vexxhost.com", "v1alpha1", "service")
def delete_fn(body, **kwargs):
    conn = openstack.connect(cloud="envvars")
    conn.identity.delete_service(
        conn.identity.find_service(body["metadata"]["name"]),
        ignore_missing=True
    )
