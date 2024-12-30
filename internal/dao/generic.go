// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Generic represents a generic resource.
type Generic struct {
	NonResource
}

// List returns a collection of resources.
func (g *Generic) List(ctx context.Context) ([]runtime.Object, error) {
	labelSel, _ := ctx.Value(internal.KeyLabels).(string)
	ns, ok := ctx.Value(internal.KeyNamespace).(string)
	if !ok {
		return nil, fmt.Errorf("BOOM!! no namespace found in context %s", g.gvr)
	}
	if client.IsAllNamespace(ns) {
		ns = client.AllNamespaces
	}

	var (
		ll  *unstructured.UnstructuredList
		err error
	)
	dial, err := g.dynClient()
	if err != nil {
		return nil, err
	}

	if client.IsClusterScoped(ns) {
		ll, err = dial.List(ctx, metav1.ListOptions{LabelSelector: labelSel})
	} else {
		ll, err = dial.Namespace(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSel})
	}
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ll.Items))
	for i := range ll.Items {
		oo[i] = &ll.Items[i]
	}

	return oo, nil
}

// Get returns a given resource.
func (g *Generic) Get(ctx context.Context, path string) (runtime.Object, error) {
	var opts metav1.GetOptions

	ns, n := client.Namespaced(path)
	dial, err := g.dynClient()
	if err != nil {
		return nil, err
	}
	if client.IsClusterScoped(ns) {
		return dial.Get(ctx, n, opts)
	}

	return dial.Namespace(ns).Get(ctx, n, opts)
}

func (g *Generic) dynClient() (dynamic.NamespaceableResourceInterface, error) {
	dial, err := g.Client().DynDial()
	if err != nil {
		return nil, err
	}

	return dial.Resource(g.gvr.GVR()), nil
}
