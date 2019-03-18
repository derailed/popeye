package k8s

//go:generate popeye gen

import (
	"k8s.io/client-go/kubernetes"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
)

var (
	supportedMetricsAPIVersions = []string{"v1beta1"}
)

// Server represents a Kubernetes api server.
type Server struct {
	config *Config

	client kubernetes.Interface
}

// NewServer returns a dialable api server configuration.
func NewServer(config *Config) *Server {
	return &Server{config: config}
}

// Dial returns a handle to api server.
func (s *Server) Dial() (kubernetes.Interface, error) {
	if s.client != nil {
		return s.client, nil
	}

	var err error
	restCfg, err := s.config.RESTConfig()
	if err != nil {
		return nil, nil
	}

	if s.client, err = kubernetes.NewForConfig(restCfg); err != nil {
		return nil, err
	}
	return s.client, nil
}

// ClusterHasMetrics checks if metrics server is on the cluster or not.
func (s *Server) ClusterHasMetrics() bool {
	srv, err := s.Dial()
	if err != nil {
		return false
	}
	apiGroups, err := srv.Discovery().ServerGroups()
	if err != nil {
		return false
	}

	for _, discoveredAPIGroup := range apiGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}
