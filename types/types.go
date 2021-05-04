package types

import (
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// CreateVerb represents create access on a resource.
	CreateVerb = "create"

	// UpdateVerb represents an update access on a resource.
	UpdateVerb = "update"

	// PatchVerb represents a patch access on a resource.
	PatchVerb = "patch"

	// DeleteVerb represents a delete access on a resource.
	DeleteVerb = "delete"

	// GetVerb represents a get access on a resource.
	GetVerb = "get"

	// ListVerb represents a list access on a resource.
	ListVerb = "list"

	// WatchVerb represents a watch access on a resource.
	WatchVerb = "watch"
)

var (
	// GetAccess reads a resource.
	GetAccess = []string{GetVerb}
	// ListAccess list resources.
	ListAccess = []string{ListVerb}
	// MonitorAccess monitors a collection of resources.
	MonitorAccess = []string{ListVerb, WatchVerb}
	// ReadAllAccess represents an all read access to a resource.
	ReadAllAccess = []string{GetVerb, ListVerb, WatchVerb}
)

// Authorizer checks what a user can or cannot do to a resource.
type Authorizer interface {
	// CanI returns true if the user can use these actions for a given resource.
	CanI(ns, gvr string, verbs []string) (bool, error)
}

// Config represents an api server configuration.
type Config interface {
	CurrentNamespaceName() (string, error)
	CurrentClusterName() (string, error)
	Flags() *genericclioptions.ConfigFlags
	RESTConfig() (*restclient.Config, error)
	CallTimeout() time.Duration
}

// Connection represents a Kubenetes apiserver connection.
type Connection interface {
	Authorizer

	// Config returns current config.
	Config() Config

	// Dial connects to api server.
	Dial() (kubernetes.Interface, error)

	// CachedDiscovery connects to discovery client.
	CachedDiscovery() (*disk.CachedDiscoveryClient, error)

	// RestConfig connects to rest client.
	RestConfig() (*restclient.Config, error)

	// MXDial connects to metrics server.
	MXDial() (*versioned.Clientset, error)

	// DynDial connects to dynamic client.
	DynDial() (dynamic.Interface, error)

	// HasMetrics checks if metrics server is available.
	HasMetrics() bool

	// ServerVersion returns current server version.
	ServerVersion() (*version.Info, error)

	// ActiveCluster returns the current cluster name.
	ActiveCluster() string

	// ActiveNamespace returns the current namespace.
	ActiveNamespace() string

	// // IsActiveNamespace checks if given ns is active.
	// IsActiveNamespace(string) bool
}

// Factory represents a resource factory.
type Factory interface {
	// Client retrieves an api client.
	Client() Connection

	// Get fetch a given resource.
	Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error)

	// List fetch a collection of resources.
	List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error)

	// ForResource fetch an informer for a given resource.
	ForResource(ns, gvr string) (informers.GenericInformer, error)

	// CanForResource fetch an informer for a given resource if authorized
	CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error)

	// WaitForCacheSync synchronize the cache.
	WaitForCacheSync()
}
