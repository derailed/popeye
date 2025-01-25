// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal

// ContextKey represents context key.
type ContextKey string

// A collection of context keys.
const (
	KeyFactory       ContextKey = "factory"
	KeyLabels        ContextKey = "labels"
	KeyFields        ContextKey = "fields"
	KeyOverAllocs    ContextKey = "overAllocs"
	KeyRunInfo       ContextKey = "runInfo"
	KeyConfig        ContextKey = "config"
	KeyNamespace     ContextKey = "namespace"
	KeyVersion       ContextKey = "version"
	KeyDB            ContextKey = "db"
	KeyNamespaceName ContextKey = "namespaceName"
)
