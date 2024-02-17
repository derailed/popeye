// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package db

import "k8s.io/apimachinery/pkg/runtime"

type ConvertFn func(o runtime.Object) (any, error)
