// Copyright (c) 2024 VEXXHOST, Inc.
// SPDX-License-Identifier: Apache-2.0

package api

import "errors"

var (
	ErrorNotFound      = errors.New("resource not found")
	ErrorMultipleFound = errors.New("multiple resources found")
)
