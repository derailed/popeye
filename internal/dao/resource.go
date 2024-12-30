// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Resource)(nil)

// Resource represents an informer based resource.
type Resource struct {
	Generic
}

// List returns a collection of resources.
func (r *Resource) List(ctx context.Context) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}
	ns, ok := ctx.Value(internal.KeyNamespace).(string)
	if !ok {
		return nil, fmt.Errorf("BOOM!! no namespace found in context %s", r.gvr)
	}
	if r.gvr == internal.Glossary[internal.NS] {
		ns = client.AllNamespaces
	}

	return r.Factory.List(r.gvr, ns, true, lsel)
}

// Get returns a resource instance if found, else an error.
func (r *Resource) Get(_ context.Context, path string) (runtime.Object, error) {
	return r.Factory.Get(r.gvr, path, true, labels.Everything())
}
