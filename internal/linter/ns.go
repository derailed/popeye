package linter

import (
	"context"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

// Namespace represents a Namespace linter.
type Namespace struct {
	*Linter
}

// NewNamespace returns a new namespace linter.
func NewNamespace(c *k8s.Client, l *zerolog.Logger) *Namespace {
	return &Namespace{newLinter(c, l)}
}

// Lint a namespace
func (n *Namespace) Lint(ctx context.Context) error {
	ll, err := n.client.ListNS()
	if err != nil {
		return err
	}
	n.lint(ll)

	return nil
}

func (n *Namespace) lint(nn []v1.Namespace) {
	for _, ns := range nn {
		n.initIssues(ns.Name)
		n.checkActive(ns)
		if ns.Status.Phase == v1.NamespaceActive {
		}
	}
}

func (n *Namespace) checkActive(ns v1.Namespace) {
	if ns.Status.Phase != v1.NamespaceActive {
		n.addIssuef(ns.Name, ErrorLevel, "Namespace is inactive")
	}
}
