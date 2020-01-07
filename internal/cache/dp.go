package cache

import (
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentKey tracks Deployment resource references
const DeploymentKey = "dp"

// Deployment represents Deployment cache.
type Deployment struct {
	dps map[string]*appsv1.Deployment
}

// NewDeployment returns a new Deployment cache.
func NewDeployment(dps map[string]*appsv1.Deployment) *Deployment {
	return &Deployment{dps: dps}
}

// ListDeployments returns all available Deployments on the cluster.
func (d *Deployment) ListDeployments() map[string]*appsv1.Deployment {
	return d.dps
}
