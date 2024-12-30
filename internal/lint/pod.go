// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// Pod represents a Pod linter.
	Pod struct {
		*issues.Collector

		db *db.DB
	}

	// PodMetric tracks pod metrics available and current range.
	PodMetric interface {
		CurrentCPU() int64
		CurrentMEM() int64
		Empty() bool
	}
)

// NewPod returns a new instance.
func NewPod(co *issues.Collector, db *db.DB) *Pod {
	return &Pod{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource..
func (s *Pod) Lint(ctx context.Context) error {
	boundSA := boundDefaultSA(s.db)
	txn, it := s.db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		po := o.(*v1.Pod)
		fqn := client.FQN(po.Namespace, po.Name)
		s.InitOutcome(fqn)
		defer s.CloseOutcome(ctx, fqn, nil)

		ctx = internal.WithSpec(ctx, coSpecFor(fqn, po, po.Spec))
		s.checkStatus(ctx, po)
		s.checkContainerStatus(ctx, fqn, po)
		s.checkContainers(ctx, fqn, po)
		s.checkOwnedByAnything(ctx, po.OwnerReferences)
		s.checkNPs(ctx, po)
		if !ownedByDaemonSet(po) {
			s.checkPdb(ctx, po.ObjectMeta.Labels)
		}
		s.checkForMultiplePdbMatches(ctx, po.Namespace, po.ObjectMeta.Labels)
		s.checkSecure(ctx, fqn, po.Spec, boundSA)

		pmx, err := s.db.FindPMX(fqn)
		if err != nil {
			continue
		}
		cmx := make(client.ContainerMetrics)
		containerMetrics(pmx, cmx)
		s.checkUtilization(ctx, fqn, po, cmx)
	}

	return nil
}

func ownedByDaemonSet(po *v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func (s *Pod) checkNPs(ctx context.Context, pod *v1.Pod) {
	txn, it := s.db.MustITForNS(internal.Glossary[internal.NP], pod.Namespace)
	defer txn.Abort()

	const (
		in  = 0
		out = 1
	)
	matches := [2]int{}
	for o := it.Next(); o != nil; o = it.Next() {
		np := o.(*netv1.NetworkPolicy)
		if isDenyAll(&np.Spec) || isAllowAll(&np.Spec) {
			return
		}
		if isDenyAllIngress(&np.Spec) || isAllowAllIngress(&np.Spec) {
			matches[in]++
			if s.checkEgresses(ctx, pod, np.Spec.Egress) {
				matches[out]++
			}
			continue
		}
		if isDenyAllEgress(&np.Spec) || isAllowAllEgress(&np.Spec) {
			matches[out]++
			if s.checkIngresses(ctx, pod, np.Spec.Ingress) {
				matches[in]++
			}
			continue
		}
		if labelsMatch(&np.Spec.PodSelector, pod.Labels) {
			if polInclude(np.Spec.PolicyTypes, dirIn) {
				matches[in]++
			}
			if polInclude(np.Spec.PolicyTypes, dirOut) {
				matches[out]++
			}
		} else {
			if s.checkIngresses(ctx, pod, np.Spec.Ingress) {
				matches[out]++
			}
			if s.checkEgresses(ctx, pod, np.Spec.Egress) {
				matches[in]++
			}
		}
	}

	if matches[in] == 0 {
		s.AddCode(ctx, 1204, ingress)
	}
	if matches[out] == 0 {
		s.AddCode(ctx, 1204, egress)
	}
}

type Labels map[string]string

func (s *Pod) isPodTargeted(pod *v1.Pod, nsSel, podSel *metav1.LabelSelector, b *netv1.IPBlock) (bool, error) {
	nn, err := s.db.FindNSNameBySel(nsSel)
	if err != nil {
		return false, err
	}
	if len(nn) == 0 && b == nil {
		if podSel == nil {
			return false, nil
		}
		return labelsMatch(podSel, pod.Labels), nil
	}
	for _, sns := range nn {
		if sns != pod.Namespace {
			continue
		}
		if podSel != nil && podSel.Size() > 0 {
			if labelsMatch(podSel, pod.Labels) {
				return true, nil
			}
		}
	}

	if b == nil {
		return false, nil
	}
	_, ipnet, err := net.ParseCIDR(b.CIDR)
	if err != nil {
		return false, err
	}
	for _, ip := range pod.Status.PodIPs {
		if ipnet.Contains(net.ParseIP(ip.IP)) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Pod) checkIngresses(ctx context.Context, pod *v1.Pod, rr []netv1.NetworkPolicyIngressRule) bool {
	var match int
	if rr == nil {
		return false
	}
	for _, r := range rr {
		if r.From == nil {
			return true
		}
		for _, from := range r.From {
			ok, err := s.isPodTargeted(pod, from.NamespaceSelector, from.PodSelector, from.IPBlock)
			if err != nil {
				s.AddErr(ctx, err)
				return true
			}
			if ok {
				match++
			}
		}
	}

	return match > 0
}

func (s *Pod) checkEgresses(ctx context.Context, pod *v1.Pod, rr []netv1.NetworkPolicyEgressRule) bool {
	if rr == nil {
		return false
	}
	var match int
	for _, r := range rr {
		if r.To == nil {
			return true
		}
		for _, to := range r.To {
			ok, err := s.isPodTargeted(pod, to.NamespaceSelector, to.PodSelector, to.IPBlock)
			if err != nil {
				s.AddErr(ctx, err)
				return true
			}
			if ok {
				match++
			}
		}
	}

	return match > 0
}

func labelsMatch(sel *metav1.LabelSelector, ll map[string]string) bool {
	if sel == nil || sel.Size() == 0 {
		return true
	}

	return db.MatchSelector(ll, sel)
}

func (s *Pod) checkOwnedByAnything(ctx context.Context, ownerRefs []metav1.OwnerReference) {
	if len(ownerRefs) == 0 {
		s.AddCode(ctx, 208)
		return
	}

	controlled := false
	for _, or := range ownerRefs {
		if or.Controller != nil && *or.Controller {
			controlled = true
			break
		}
	}

	if !controlled {
		s.AddCode(ctx, 208)
	}
}

func (s *Pod) checkPdb(ctx context.Context, labels map[string]string) {
	if s.ForLabels(labels) == nil {
		s.AddCode(ctx, 206)
	}
}

// ForLabels returns a pdb whose selector match the given labels. Returns nil if no match.
func (s *Pod) ForLabels(labels map[string]string) *policyv1.PodDisruptionBudget {
	txn, it := s.db.MustITFor(internal.Glossary[internal.PDB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pdb := o.(*policyv1.PodDisruptionBudget)
		m, err := metav1.LabelSelectorAsMap(pdb.Spec.Selector)
		if err != nil {
			continue
		}
		if cache.MatchLabels(labels, m) {
			return pdb
		}
	}
	return nil
}

func (s *Pod) checkUtilization(ctx context.Context, fqn string, po *v1.Pod, cmx client.ContainerMetrics) {
	if len(cmx) == 0 {
		return
	}
	for _, co := range po.Spec.Containers {
		cmx, ok := cmx[co.Name]
		if !ok {
			continue
		}
		NewContainer(fqn, s).checkUtilization(ctx, co, cmx)
	}
}

func (s *Pod) checkSecure(ctx context.Context, fqn string, spec v1.PodSpec, boundSA bool) {
	if err := s.checkSA(ctx, fqn, spec, boundSA); err != nil {
		s.AddErr(ctx, err)
	}
	s.checkSecContext(ctx, fqn, spec)
}

func (s *Pod) checkSA(ctx context.Context, fqn string, spec v1.PodSpec, boundSA bool) error {
	ns, _ := namespaced(fqn)
	if spec.ServiceAccountName == "default" {
		if boundSA {
			s.AddCode(ctx, 308)
		} else {
			s.AddCode(ctx, 300)
		}
	}

	txn := s.db.Txn(false)
	defer txn.Abort()
	saFQN := cache.FQN(ns, spec.ServiceAccountName)
	o, err := txn.First(internal.Glossary[internal.SA].String(), "id", saFQN)
	if err != nil || o == nil {
		s.AddCode(ctx, 307, "Pod", spec.ServiceAccountName)
		if isBoolSet(spec.AutomountServiceAccountToken) {
			s.AddCode(ctx, 301)
		}
		return nil
	}
	sa, ok := o.(*v1.ServiceAccount)
	if !ok {
		return fmt.Errorf("expecting SA %q but got %T", saFQN, o)
	}
	if spec.AutomountServiceAccountToken == nil {
		if isBoolSet(sa.AutomountServiceAccountToken) {
			s.AddCode(ctx, 301)
		}
	} else if isBoolSet(spec.AutomountServiceAccountToken) {
		s.AddCode(ctx, 301)
	}

	return nil
}

func (s *Pod) checkSecContext(ctx context.Context, fqn string, spec v1.PodSpec) {
	if spec.SecurityContext == nil {
		return
	}

	// If pod security ctx is present and we have
	podSec := hasPodNonRootUser(spec.SecurityContext)
	var victims int
	for _, co := range spec.InitContainers {
		if !checkCOSecurityContext(co) && !podSec {
			victims++
			s.AddSubCode(internal.WithGroup(ctx, types.NewGVR("containers"), co.Name), 306)
		}
	}
	for _, co := range spec.Containers {
		if !checkCOSecurityContext(co) && !podSec {
			victims++
			s.AddSubCode(internal.WithGroup(ctx, types.NewGVR("containers"), co.Name), 306)
		}
	}
	if victims > 0 && !podSec {
		s.AddCode(ctx, 302)
	}
}

func checkCOSecurityContext(co v1.Container) bool {
	return hasCoNonRootUser(co.SecurityContext)
}

func hasPodNonRootUser(sec *v1.PodSecurityContext) bool {
	if sec == nil {
		return false
	}
	if sec.RunAsNonRoot != nil {
		return *sec.RunAsNonRoot
	}
	if sec.RunAsUser != nil {
		return *sec.RunAsUser != 0
	}
	return false
}

func hasCoNonRootUser(sec *v1.SecurityContext) bool {
	if sec == nil {
		return false
	}
	if sec.RunAsNonRoot != nil {
		return *sec.RunAsNonRoot
	}
	if sec.RunAsUser != nil {
		return *sec.RunAsUser != 0
	}
	return false
}

func (s *Pod) checkContainers(ctx context.Context, fqn string, po *v1.Pod) {
	co := NewContainer(fqn, s)
	for _, c := range po.Spec.InitContainers {
		co.sanitize(ctx, c, false)
	}
	for _, c := range po.Spec.Containers {
		co.sanitize(ctx, c, !isPartOfJob(po))
	}
}

func (s *Pod) checkContainerStatus(ctx context.Context, fqn string, po *v1.Pod) {
	limit := s.RestartsLimit()
	size := len(po.Status.InitContainerStatuses)
	for _, cs := range po.Status.InitContainerStatuses {
		newContainerStatus(s, fqn, size, true, limit).sanitize(ctx, cs)
	}

	size = len(po.Status.ContainerStatuses)
	for _, cs := range po.Status.ContainerStatuses {
		newContainerStatus(s, fqn, size, false, limit).sanitize(ctx, cs)
	}
}

func (s *Pod) checkStatus(ctx context.Context, po *v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		s.AddCode(ctx, 207, po.Status.Phase)
	}
}

// !!BOZO!! Check
func (s *Pod) checkForMultiplePdbMatches(ctx context.Context, podNamespace string, podLabels map[string]string) {
	matchedPdbs := make([]string, 0, 10)
	txn, it := s.db.MustITFor(internal.Glossary[internal.PDB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		pdb := o.(*policyv1.PodDisruptionBudget)
		if podNamespace != pdb.Namespace {
			continue
		}
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			log.Error().Err(err).Msg("No selectors found")
			return
		}
		if selector.Empty() || !selector.Matches(labels.Set(podLabels)) {
			continue
		}
		matchedPdbs = append(matchedPdbs, pdb.Name)
	}
	if len(matchedPdbs) > 1 {
		sort.Strings(matchedPdbs)
		s.AddCode(ctx, 209, strings.Join(matchedPdbs, ", "))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func containerMetrics(pmx *mv1beta1.PodMetrics, mx client.ContainerMetrics) {
	if pmx == nil {
		return
	}

	for _, co := range pmx.Containers {
		mx[co.Name] = client.Metrics{
			CurrentCPU: *co.Usage.Cpu(),
			CurrentMEM: *co.Usage.Memory(),
		}
	}
}

func isPartOfJob(po *v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "Job" {
			return true
		}
	}

	return false
}

func isBoolSet(b *bool) bool {
	return b != nil && *b
}
