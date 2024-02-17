// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/types"
	"k8s.io/apimachinery/pkg/runtime"
)

// NonResource represents a non k8s resource.
type NonResource struct {
	types.Factory

	gvr types.GVR
}

// Init initializes the resource.
func (n *NonResource) Init(f types.Factory, gvr types.GVR) {
	n.Factory, n.gvr = f, gvr
}

// GVR returns a gvr.
func (n *NonResource) GVR() string {
	return n.gvr.String()
}

// Get returns the given resource.
func (n *NonResource) Get(context.Context, string) (runtime.Object, error) {
	return nil, fmt.Errorf("NYI!")
}
