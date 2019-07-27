package cache

import (
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentKey tracks Deployment ressource references
const DeploymentKey = "dp"

// Deployment represents Deployment cache.
type Deployment struct {
	dps          map[string]*appsv1.Deployment
	preferredRev string
}

// NewDeployment returns a new Deployment cache.
func NewDeployment(dps map[string]*appsv1.Deployment, rev string) *Deployment {
	return &Deployment{dps: dps, preferredRev: rev}
}

// ListDeployments returns all available Deployments on the cluster.
func (d *Deployment) ListDeployments() map[string]*appsv1.Deployment {
	return d.dps
}

// DeploymentPreferredRev return API server preferred rev.
func (d *Deployment) DeploymentPreferredRev() string {
	return d.preferredRev
}
