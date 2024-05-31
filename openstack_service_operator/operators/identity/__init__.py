# SPDX-License-Identifier: Apache-2.0

import kopf

from openstack_service_operator.operators.identity import endpoint
from openstack_service_operator.operators.identity import service


@kopf.on.startup()
def configure(settings: kopf.OperatorSettings, **_):
    # TODO(mnaser): Move this to root
    settings.persistence.finalizer = "openstack.vexxhost.com/finalizer"
