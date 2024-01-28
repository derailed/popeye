// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dao

import (
	"context"

	"github.com/derailed/popeye/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[types.GVR]metav1.APIResource

// Getter represents a resource getter.
type Getter interface {
	// Get return a given resource.
	Get(ctx context.Context, path string) (runtime.Object, error)
}

// Lister represents a resource lister.
type Lister interface {
	// List returns a resource collection.
	List(ctx context.Context) ([]runtime.Object, error)
}

// Accessor represents an accessible k8s resource.
type Accessor interface {
	Lister
	Getter

	// Init the resource with a factory object.
	Init(types.Factory, types.GVR)

	// GVR returns a gvr a string.
	GVR() string
}
