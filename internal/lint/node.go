// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
)

type (
	tolerations map[string]struct{}

	// Node represents a Node linter.
	Node struct {
		*issues.Collector

		db *db.DB
	}
)

// NewNode returns a new instance.
func NewNode(co *issues.Collector, db *db.DB) *Node {
	return &Node{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (n *Node) Lint(ctx context.Context) error {
	nmx := make(client.NodesMetrics)
	n.nodesMetrics(nmx)

	tt, err := n.fetchPodTolerations()
	if err != nil {
		return err
	}
	txn, it := n.db.MustITFor(internal.Glossary[internal.NO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		no := o.(*v1.Node)
		fqn := no.Name
		n.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, no))

		n.checkConditions(ctx, no)
		if err := n.checkTaints(ctx, no.Spec.Taints, tt); err != nil {
			n.AddErr(ctx, err)
		}
		n.checkUtilization(ctx, nmx[fqn])
	}

	return nil
}

func (n *Node) checkTaints(ctx context.Context, taints []v1.Taint, tt tolerations) error {
	for _, ta := range taints {
		if _, ok := tt[mkKey(ta.Key, ta.Value)]; !ok {
			n.AddCode(ctx, 700, ta.Key)
		}
	}

	return nil
}

func (n *Node) fetchPodTolerations() (tolerations, error) {
	tt := make(tolerations)
	txn, it := n.db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()

	for o := it.Next(); o != nil; o = it.Next() {
		po, ok := o.(*v1.Pod)
		if !ok {
			return nil, errors.New("po conversion failed")
		}
		for _, t := range po.Spec.Tolerations {
			tt[mkKey(t.Key, t.Value)] = struct{}{}
		}
	}

	return tt, nil
}

func mkKey(k, v string) string {
	return k + ":" + v
}

func (n *Node) checkConditions(ctx context.Context, no *v1.Node) {
	if no.Spec.Unschedulable {
		n.AddCode(ctx, 711)
	}
	for _, c := range no.Status.Conditions {
		if c.Status == v1.ConditionUnknown {
			n.AddCode(ctx, 701)
		}
		if c.Type == v1.NodeReady && c.Status == v1.ConditionFalse {
			n.AddCode(ctx, 702)
		}
		n.statusReport(ctx, c.Type, c.Status)
	}
}

func (n *Node) statusReport(ctx context.Context, cond v1.NodeConditionType, status v1.ConditionStatus) {
	if status == v1.ConditionFalse {
		return
	}

	switch cond {
	case v1.NodeMemoryPressure:
		n.AddCode(ctx, 704)
	case v1.NodeDiskPressure:
		n.AddCode(ctx, 705)
	case v1.NodePIDPressure:
		n.AddCode(ctx, 706)
	case v1.NodeNetworkUnavailable:
		n.AddCode(ctx, 707)
	}
}

func (n *Node) checkUtilization(ctx context.Context, mx client.NodeMetrics) {
	if mx.Empty() {
		n.AddCode(ctx, 708)
		return
	}

	percCPU := ToPerc(toMC(mx.CurrentCPU), toMC(mx.AvailableCPU))
	cpuLimit := int64(n.NodeCPULimit())
	if percCPU > cpuLimit {
		n.AddCode(ctx, 709, cpuLimit, percCPU)
	}

	percMEM := ToPerc(toMB(mx.CurrentMEM), toMB(mx.AvailableMEM))
	memLimit := int64(n.NodeMEMLimit())
	if percMEM > memLimit {
		n.AddCode(ctx, 710, memLimit, percMEM)
	}
}

func (n *Node) nodesMetrics(nmx client.NodesMetrics) {
	mm, err := n.db.ListNMX()
	if err != nil || len(mm) == 0 {
		return
	}

	txn, it := n.db.MustITFor(internal.Glossary[internal.NO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		no := o.(*v1.Node)
		if len(no.Status.Allocatable) == 0 && len(no.Status.Capacity) == 0 {
			continue
		}
		nmx[no.Name] = client.NodeMetrics{
			AvailableCPU: *no.Status.Allocatable.Cpu(),
			AvailableMEM: *no.Status.Allocatable.Memory(),
			TotalCPU:     *no.Status.Capacity.Cpu(),
			TotalMEM:     *no.Status.Capacity.Memory(),
		}
	}

	for _, m := range mm {
		if mx, ok := nmx[m.Name]; ok {
			mx.CurrentCPU = *m.Usage.Cpu()
			mx.CurrentMEM = *m.Usage.Memory()
			nmx[m.Name] = mx
		}
	}
}
