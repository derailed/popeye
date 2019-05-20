package cache

import (
	v1 "k8s.io/api/core/v1"
)

// Node represents a collection of Nodes available on a cluster.
type Node struct {
	nodes map[string]*v1.Node
}

// NewNode returns a new Node.
func NewNode(svcs map[string]*v1.Node) *Node {
	return &Node{svcs}
}

// ListNodes returns all available Nodes on the cluster.
func (n *Node) ListNodes() map[string]*v1.Node {
	return n.nodes
}
