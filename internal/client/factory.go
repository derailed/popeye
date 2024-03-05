// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	di "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
)

const (
	defaultResync   = 10 * time.Minute
	defaultWaitTime = 250 * time.Millisecond
)

// Factory tracks various resource informers.
type Factory struct {
	factories map[string]di.DynamicSharedInformerFactory
	client    types.Connection
	stopChan  chan struct{}
	mx        sync.RWMutex
}

// NewFactory returns a new informers factory.
func NewFactory(client types.Connection) *Factory {
	return &Factory{
		client:    client,
		factories: make(map[string]di.DynamicSharedInformerFactory),
	}
}

// Start initializes the informers until caller cancels the context.
func (f *Factory) Start(ns string) {
	f.mx.Lock()
	defer f.mx.Unlock()

	log.Debug().Msgf("Factory START with ns `%q", ns)
	f.stopChan = make(chan struct{})
	for ns, fac := range f.factories {
		log.Debug().Msgf("Starting factory in ns %q", ns)
		fac.Start(f.stopChan)
	}
}

// Terminate terminates all watchers and forwards.
func (f *Factory) Terminate() {
	f.mx.Lock()
	defer f.mx.Unlock()

	if f.stopChan != nil {
		close(f.stopChan)
		f.stopChan = nil
	}
	for k := range f.factories {
		delete(f.factories, k)
	}
}

// List returns a resource collection.
func (f *Factory) List(gvr types.GVR, ns string, wait bool, labels labels.Selector) ([]runtime.Object, error) {
	inf, err := f.CanForResource(ns, gvr, types.ListAccess)
	if err != nil {
		return nil, err
	}
	if IsAllNamespace(ns) {
		ns = BlankNamespace
	}

	var oo []runtime.Object
	if IsClusterScoped(ns) {
		oo, err = inf.Lister().List(labels)
	} else {
		oo, err = inf.Lister().ByNamespace(ns).List(labels)
	}
	if !wait || (wait && inf.Informer().HasSynced()) {
		return oo, err
	}

	f.waitForCacheSync(ns)
	if IsClusterScoped(ns) {
		return inf.Lister().List(labels)
	}
	return inf.Lister().ByNamespace(ns).List(labels)
}

// HasSynced checks if given informer is up to date.
func (f *Factory) HasSynced(gvr types.GVR, ns string) (bool, error) {
	inf, err := f.CanForResource(ns, gvr, types.ListAccess)
	if err != nil {
		return false, err
	}

	return inf.Informer().HasSynced(), nil
}

// Get retrieves a given resource.
func (f *Factory) Get(gvr types.GVR, fqn string, wait bool, sel labels.Selector) (runtime.Object, error) {
	ns, n := Namespaced(fqn)
	inf, err := f.CanForResource(ns, gvr, []string{types.GetVerb})
	if err != nil {
		return nil, err
	}
	var o runtime.Object
	if IsClusterScoped(ns) {
		o, err = inf.Lister().Get(n)
	} else {
		o, err = inf.Lister().ByNamespace(ns).Get(n)
	}
	if !wait || (wait && inf.Informer().HasSynced()) {
		return o, err
	}

	f.waitForCacheSync(ns)
	if IsClusterScoped(ns) {
		return inf.Lister().Get(n)
	}
	return inf.Lister().ByNamespace(ns).Get(n)
}

func (f *Factory) waitForCacheSync(ns string) {
	if IsClusterWide(ns) {
		ns = BlankNamespace
	}

	f.mx.RLock()
	defer f.mx.RUnlock()
	fac, ok := f.factories[ns]
	if !ok {
		return
	}

	// Hang for a sec for the cache to refresh if still not done bail out!
	c := make(chan struct{})
	go func(c chan struct{}) {
		<-time.After(defaultWaitTime)
		close(c)
	}(c)
	_ = fac.WaitForCacheSync(c)
}

// WaitForCacheSync waits for all factories to update their cache.
func (f *Factory) WaitForCacheSync() {
	for ns, fac := range f.factories {
		m := fac.WaitForCacheSync(f.stopChan)
		for k, v := range m {
			log.Debug().Msgf("CACHE `%q Loaded %t:%s", ns, v, k)
		}
	}
}

// Client return the factory connection.
func (f *Factory) Client() types.Connection {
	return f.client
}

// FactoryFor returns a factory for a given namespace.
func (f *Factory) FactoryFor(ns string) di.DynamicSharedInformerFactory {
	return f.factories[ns]
}

// SetActiveNS sets the active namespace.
func (f *Factory) SetActiveNS(ns string) error {
	if f.isClusterWide() {
		return nil
	}
	_, err := f.ensureFactory(ns)
	return err
}

func (f *Factory) isClusterWide() bool {
	f.mx.RLock()
	defer f.mx.RUnlock()
	_, ok := f.factories[BlankNamespace]

	return ok
}

// CanForResource return an informer is user has access.
func (f *Factory) CanForResource(ns string, gvr types.GVR, verbs []string) (informers.GenericInformer, error) {
	auth, err := f.Client().CanI(ns, gvr, "", verbs)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("%v access denied on resource %q:%q", verbs, ns, gvr)
	}

	return f.ForResource(ns, gvr)
}

// ForResource returns an informer for a given resource.
func (f *Factory) ForResource(ns string, gvr types.GVR) (informers.GenericInformer, error) {
	fact, err := f.ensureFactory(ns)
	if err != nil {
		return nil, err
	}
	inf := fact.ForResource(gvr.GVR())
	if inf == nil {
		log.Error().Err(fmt.Errorf("MEOW! No informer for %q:%q", ns, gvr))
		return inf, nil
	}

	f.mx.RLock()
	defer f.mx.RUnlock()
	fact.Start(f.stopChan)

	return inf, nil
}

func (f *Factory) ensureFactory(ns string) (di.DynamicSharedInformerFactory, error) {
	if IsClusterWide(ns) {
		ns = BlankNamespace
	}
	f.mx.Lock()
	defer f.mx.Unlock()
	if fac, ok := f.factories[ns]; ok {
		return fac, nil
	}

	dial, err := f.client.DynDial()
	if err != nil {
		return nil, err
	}
	f.factories[ns] = di.NewFilteredDynamicSharedInformerFactory(
		dial,
		defaultResync,
		ns,
		nil,
	)

	return f.factories[ns], nil
}
