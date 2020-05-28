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

// Terminate stops the factory.
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
func (f *Factory) List(gvr, ns string, wait bool, labels labels.Selector) ([]runtime.Object, error) {
	inf, err := f.CanForResource(ns, gvr, types.MonitorAccess)
	if err != nil {
		return nil, err
	}
	if wait {
		f.waitForCacheSync(ns)
	}
	if IsClusterScoped(ns) {
		return inf.Lister().List(labels)
	}

	if IsAllNamespace(ns) {
		ns = AllNamespaces
	}
	return inf.Lister().ByNamespace(ns).List(labels)
}

// Get retrieves a given resource.
func (f *Factory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	ns, n := Namespaced(path)
	inf, err := f.CanForResource(ns, gvr, []string{types.GetVerb})
	if err != nil {
		return nil, err
	}

	if wait {
		f.waitForCacheSync(ns)
	}
	if IsClusterScoped(ns) {
		return inf.Lister().Get(n)
	}

	return inf.Lister().ByNamespace(ns).Get(n)
}

func (f *Factory) waitForCacheSync(ns string) {
	if IsClusterWide(ns) {
		ns = AllNamespaces
	}

	if f.isClusterWide() {
		ns = AllNamespaces
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
			log.Debug().Msgf("CACHE %q synched %t:%s", ns, v, k)
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
	if !f.isClusterWide() {
		if _, err := f.ensureFactory(ns); err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) isClusterWide() bool {
	f.mx.RLock()
	defer f.mx.RUnlock()

	_, ok := f.factories[AllNamespaces]
	return ok
}

// CanForResource return an informer is user has access.
func (f *Factory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	// If user can access resource cluster wide, prefer cluster wide factory.
	if !IsClusterWide(ns) {
		auth, err := f.Client().CanI(AllNamespaces, gvr, verbs)
		if auth && err == nil {
			return f.ForResource(AllNamespaces, gvr)
		}
	}
	auth, err := f.Client().CanI(ns, gvr, verbs)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("%v access denied on resource %q:%q", verbs, ns, gvr)
	}

	return f.ForResource(ns, gvr)
}

// ForResource returns an informer for a given resource.
func (f *Factory) ForResource(ns, gvr string) (informers.GenericInformer, error) {
	fact, err := f.ensureFactory(ns)
	if err != nil {
		return nil, err
	}
	inf := fact.ForResource(NewGVR(gvr).GVR())
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
		ns = AllNamespaces
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
