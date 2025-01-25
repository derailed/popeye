// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	authorizationv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	cacheSize     = 100
	cacheExpiry   = 5 * time.Minute
	cacheMXAPIKey = "metricsAPI"
	// CallTimeout represents api call timeout limit.
	CallTimeout = 30 * time.Second
)

var supportedMetricsAPIVersions = []string{"v1beta1"}

// APIClient represents a Kubernetes api client.
type APIClient struct {
	checkClientSet *kubernetes.Clientset
	client         kubernetes.Interface
	dClient        dynamic.Interface
	mxsClient      *versioned.Clientset
	cachedClient   *disk.CachedDiscoveryClient
	config         types.Config
	mx             sync.RWMutex
	cache          *cache.LRUExpireCache
}

// NewTestClient for testing ONLY!!
func NewTestClient() *APIClient {
	return &APIClient{
		config: NewConfig(nil),
		cache:  cache.NewLRUExpireCache(cacheSize),
	}
}

// InitConnectionOrDie initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnectionOrDie(config types.Config) (*APIClient, error) {
	a := APIClient{
		config: config,
		cache:  cache.NewLRUExpireCache(cacheSize),
	}
	if _, err := a.serverGroups(); err != nil {
		return nil, fmt.Errorf("init connection fail: %w", err)
	}
	if err := a.supportsMetricsResources(); err != nil {
		log.Warn().Err(err).Msgf("no metrics server detected")
	}

	return &a, nil
}

func makeSAR(ns string, gvr types.GVR, n string) *authorizationv1.SelfSubjectAccessReview {
	if ns == "-" {
		ns = ""
	}
	res := gvr.GVR()
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   ns,
				Group:       res.Group,
				Resource:    res.Resource,
				Subresource: gvr.SubResource(),
				Name:        n,
			},
		},
	}
}

func makeCacheKey(ns string, gvr types.GVR, n string, vv []string) string {
	return ns + ":" + gvr.String() + ":" + n + "::" + strings.Join(vv, ",")
}

// ActiveContext returns the current context name.
func (a *APIClient) ActiveContext() string {
	c, err := a.config.CurrentContextName()
	if err != nil {
		log.Error().Msgf("Unable to located active context")
		return ""
	}

	return c
}

// ActiveCluster returns the current cluster name.
func (a *APIClient) ActiveCluster() string {
	cl, err := a.config.CurrentClusterName()
	if err != nil {
		log.Error().Msgf("Unable to located active cluster")
		return ""
	}

	return cl
}

// IsActiveNamespace returns true if namespaces matches.
func (a *APIClient) IsActiveNamespace(ns string) bool {
	if a.ActiveNamespace() == AllNamespaces {
		return true
	}
	return a.ActiveNamespace() == ns
}

// ActiveNamespace returns the current namespace.
func (a *APIClient) ActiveNamespace() string {
	ns, err := a.CurrentNamespaceName()
	if err != nil {
		return DefaultNamespace
	}

	return ns
}

func (a *APIClient) clearCache() {
	for _, k := range a.cache.Keys() {
		a.cache.Remove(k)
	}
}

// ConnectionOK checks api server connection status.
func (a *APIClient) ConnectionOK() bool {
	_, err := a.Dial()

	return err == nil
}

// CanI checks if user has access to a certain resource.
func (a *APIClient) CanI(ns string, gvr types.GVR, n string, verbs []string) (auth bool, err error) {
	if IsClusterWide(ns) {
		ns = BlankNamespace
	}
	key := makeCacheKey(ns, gvr, n, verbs)
	if v, ok := a.cache.Get(key); ok {
		if auth, ok = v.(bool); ok {
			return auth, nil
		}
	}

	c, err := a.Dial()
	if err != nil {
		return false, err
	}
	dial, sar := c.AuthorizationV1().SelfSubjectAccessReviews(), makeSAR(ns, gvr, n)
	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	defer cancel()
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := dial.Create(ctx, sar, metav1.CreateOptions{})
		if err != nil {
			log.Warn().Err(err).Msgf("  Dial Failed!")
			a.cache.Add(key, false, cacheExpiry)
			return auth, err
		}
		if !resp.Status.Allowed {
			a.cache.Add(key, false, cacheExpiry)
			return auth, fmt.Errorf("`%s access denied for user on %q:%s", v, ns, gvr)
		}
	}

	auth = true
	a.cache.Add(key, true, cacheExpiry)
	return
}

// CurrentNamespaceName return namespace name set via either cli arg or cluster config.
func (a *APIClient) CurrentNamespaceName() (string, error) {
	return a.config.CurrentNamespaceName()
}

// ServerVersion returns the current server version info.
func (a *APIClient) ServerVersion() (*version.Info, error) {
	cfg, err := a.CachedDiscovery()
	if err != nil {
		return nil, err
	}

	return cfg.ServerVersion()
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	defer cancel()

	dial, err := a.Dial()
	if err != nil {
		return nil, err
	}
	nn, err := dial.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nn.Items, nil
}

// CheckConnectivity return true if api server is cool or false otherwise.
func (a *APIClient) CheckConnectivity() (status bool) {
	defer func() {
		if !status {
			a.clearCache()
		}
		if err := recover(); err != nil {
			status = false
		}
	}()

	if a.checkClientSet == nil {
		cfg, err := a.config.Flags().ToRESTConfig()
		if err != nil {
			return
		}
		cfg.Timeout = defaultCallTimeoutDuration

		if a.checkClientSet, err = kubernetes.NewForConfig(cfg); err != nil {
			log.Error().Err(err).Msgf("Unable to connect to api server")
			return
		}
	}

	if _, err := a.checkClientSet.ServerVersion(); err == nil {
		status = true
	} else {
		log.Error().Err(err).Msgf("K9s can't connect to cluster")
	}

	return
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() types.Config {
	return a.config
}

// HasMetrics checks if the cluster supports metrics and user is authorized to use metrics.
func (a *APIClient) HasMetrics() bool {
	return a.supportsMetricsResources() == nil
}

// Dial returns a handle to api server or an error
func (a *APIClient) Dial() (kubernetes.Interface, error) {
	a.mx.Lock()
	defer a.mx.Unlock()
	if a.client != nil {
		return a.client, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if a.client, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}
	return a.client, nil
}

// RestConfig returns a rest api client.
func (a *APIClient) RestConfig() (*restclient.Config, error) {
	cfg, err := a.config.RESTConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// CachedDiscovery returns a cached discovery client.
func (a *APIClient) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.cachedClient != nil {
		return a.cachedClient, nil
	}

	rc, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(rc.Host))

	a.cachedClient, err = disk.NewCachedDiscoveryClientForConfig(rc, discCacheDir, httpCacheDir, 10*time.Minute)
	if err != nil {
		log.Panic().Msgf("Unable to connect to discovery client %v", err)
	}
	return a.cachedClient, nil
}

// DynDial returns a handle to a dynamic interface.
func (a *APIClient) DynDial() (dynamic.Interface, error) {
	a.mx.RLock()
	if a.dClient != nil {
		a.mx.RUnlock()
		return a.dClient, nil
	}
	a.mx.RUnlock()

	a.mx.Lock()
	defer a.mx.Unlock()
	rc, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if a.dClient, err = dynamic.NewForConfig(rc); err != nil {
		return nil, err
	}
	return a.dClient, nil
}

// MXDial returns a handle to the metrics server.
func (a *APIClient) MXDial() (*versioned.Clientset, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.mxsClient != nil {
		return a.mxsClient, nil
	}
	rc, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if a.mxsClient, err = versioned.NewForConfig(rc); err != nil {
		log.Error().Err(err)
	}

	return a.mxsClient, err
}

func (a *APIClient) checkCacheBool(key string) (state bool, ok bool) {
	v, found := a.cache.Get(key)
	if !found {
		return
	}
	state, ok = v.(bool)
	return
}

func (a *APIClient) serverGroups() (*metav1.APIGroupList, error) {
	dial, err := a.CachedDiscovery()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to dial discovery API")
		return nil, fmt.Errorf("unable to dial discovery: %w", err)
	}
	apiGroups, err := dial.ServerGroups()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to retrieve server groups")
		return nil, fmt.Errorf("unable to fetch server groups: %w", err)
	}

	return apiGroups, nil
}

func (a *APIClient) supportsMetricsResources() error {
	supported, ok := a.checkCacheBool(cacheMXAPIKey)
	if ok {
		if supported {
			return nil
		}
		return errors.New("no metrics-server detected")
	}
	defer func() {
		a.cache.Add(cacheMXAPIKey, supported, cacheExpiry)
	}()

	gg, err := a.serverGroups()
	if err != nil {
		return fmt.Errorf("supportmetricsResources call fail: %w", err)
	}
	for _, grp := range gg.Groups {
		if grp.Name != metricsapi.GroupName {
			continue
		}
		if checkMetricsVersion(grp) {
			supported = true
			return nil
		}
	}

	return errors.New("no metrics-server detected")
}

func checkMetricsVersion(grp metav1.APIGroup) bool {
	for _, version := range grp.Versions {
		for _, supportedVersion := range supportedMetricsAPIVersions {
			if version.Version == supportedVersion {
				return true
			}
		}
	}

	return false
}
