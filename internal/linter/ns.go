package linter

import (
	"context"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

var specialNS = []string{"default", "kube-public", "kube-node-lease"}

// Namespace represents a Namespace linter.
type Namespace struct {
	*Linter
}

// NewNamespace returns a new namespace linter.
func NewNamespace(c Client, l *zerolog.Logger) *Namespace {
	return &Namespace{newLinter(c, l)}
}

// Lint a namespace
func (n *Namespace) Lint(ctx context.Context) error {
	available, err := n.client.ListNS()
	if err != nil {
		return err
	}

	used := make([]string, len(available))
	n.client.InUseNamespaces(used)
	n.lint(available, used)

	return nil
}

func (n *Namespace) lint(nn map[string]v1.Namespace, used []string) {
	for _, ns := range nn {
		if n.client.ExcludedNS(ns.Name) {
			continue
		}
		n.initIssues(ns.Name)
		if n.checkActive(ns) {
			n.checkInUse(ns.Name, used)
		}
	}
}

func (n *Namespace) checkActive(ns v1.Namespace) bool {
	if ns.Status.Phase != v1.NamespaceActive {
		n.addIssuef(ns.Name, ErrorLevel, "Namespace is inactive")
		return false
	}
	return true
}

func (n *Namespace) checkInUse(name string, used []string) {
	if len(used) == 0 {
		return
	}

	for _, ns := range used {
		if ns == name {
			return
		}
	}

	if !in(specialNS, name) {
		n.addIssuef(name, InfoLevel, "Used?")
	}
}
