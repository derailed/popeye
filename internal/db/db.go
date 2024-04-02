// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package db

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db/schema"
	"github.com/derailed/popeye/types"
	"github.com/hashicorp/go-memdb"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type DB struct {
	*memdb.MemDB
}

func NewDB(db *memdb.MemDB) *DB {
	return &DB{
		MemDB: db,
	}
}

func (db *DB) ITFor(gvr types.GVR) (*memdb.Txn, memdb.ResultIterator, error) {
	if gvr == types.BlankGVR {
		return nil, nil, fmt.Errorf("invalid table")
	}

	txn := db.Txn(false)
	it, err := txn.Get(gvr.String(), "id")
	if err != nil {
		return nil, nil, err
	}

	return txn, it, nil
}

func (db *DB) MustITForNS(gvr types.GVR, ns string) (*memdb.Txn, memdb.ResultIterator) {
	txn := db.Txn(false)
	it, err := txn.Get(gvr.String(), "ns", ns)
	if err != nil {
		panic(fmt.Errorf("db ns iterator failed for %q: %w", gvr, err))
	}

	return txn, it
}

func (db *DB) MustITFor(gvr types.GVR) (*memdb.Txn, memdb.ResultIterator) {
	txn := db.Txn(false)
	it, err := txn.Get(gvr.String(), "id")
	if err != nil {
		panic(fmt.Errorf("db iterator failed for %q: %w", gvr, err))
	}

	return txn, it
}

func (db *DB) ListNodes() (map[string]*v1.Node, error) {
	txn, it := db.MustITFor(internal.Glossary[internal.NO])
	defer txn.Abort()

	mm := make(map[string]*v1.Node)
	for o := it.Next(); o != nil; o = it.Next() {
		no, ok := o.(*v1.Node)
		if !ok {
			return nil, fmt.Errorf("expecting node but got %T", o)
		}
		mm[no.Name] = no
	}

	return mm, nil
}

func (db *DB) FindPMX(fqn string) (*mv1beta1.PodMetrics, error) {
	gvr := internal.Glossary[internal.PMX]
	if gvr == types.BlankGVR {
		return nil, nil
	}
	txn := db.Txn(false)
	defer txn.Abort()
	o, err := txn.First(gvr.String(), "id", fqn)
	if err != nil || o == nil {
		return nil, fmt.Errorf("object not found: %q", fqn)
	}

	pmx, ok := o.(*mv1beta1.PodMetrics)
	if !ok {
		return nil, fmt.Errorf("expecting PodMetrics but got %T", o)
	}

	return pmx, nil
}

func (db *DB) FindNMX(fqn string) (*mv1beta1.NodeMetrics, error) {
	gvr := internal.Glossary[internal.NMX]
	if gvr == types.BlankGVR {
		return nil, nil
	}
	txn := db.Txn(false)
	defer txn.Abort()
	o, err := txn.First(gvr.String(), "id", fqn)
	if err != nil || o == nil {
		return nil, fmt.Errorf("object not found: %q", fqn)
	}

	nmx, ok := o.(*mv1beta1.NodeMetrics)
	if !ok {
		return nil, fmt.Errorf("expecting NodeMetrics but got %T", o)
	}

	return nmx, nil
}

func (db *DB) ListNMX() ([]*mv1beta1.NodeMetrics, error) {
	gvr := internal.Glossary[internal.NMX]
	if gvr == types.BlankGVR {
		return nil, nil
	}
	txn, it, err := db.ITFor(gvr)
	if err != nil {
		return nil, err
	}
	defer txn.Abort()

	mm := make([]*mv1beta1.NodeMetrics, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		nmx, ok := o.(*mv1beta1.NodeMetrics)
		if !ok {
			return nil, fmt.Errorf("expecting NodeMetrics but got %T", o)
		}
		mm = append(mm, nmx)
	}

	return mm, nil
}

func (db *DB) Find(kind types.GVR, fqn string) (any, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	o, err := txn.First(kind.String(), "id", fqn)
	if err != nil || o == nil {
		log.Error().Err(err).Msgf("db.find unable to find object: [%s]%s", kind, fqn)
		return nil, fmt.Errorf("object not found: %q", fqn)
	}

	return o, nil
}

func (db *DB) Dump(gvr types.GVR) {
	txn, it := db.MustITFor(gvr)
	defer txn.Abort()

	log.Debug().Msgf("> Dumping %q", gvr)
	for o := it.Next(); o != nil; o = it.Next() {
		m := o.(schema.MetaAccessor)
		log.Debug().Msgf("  o %s/%s", m.GetNamespace(), m.GetName())
	}
	log.Debug().Msg("< Done")
}

func (db *DB) FindPod(ns string, sel map[string]string) (*v1.Pod, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		po, ok := o.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("expecting pod")
		}
		if po.Namespace != ns {
			continue
		}
		if MatchLabels(po.Labels, sel) {
			return po, nil
		}
	}

	return nil, fmt.Errorf("no pods match selector: %v", sel)
}

func (db *DB) FindJobs(fqn string) ([]*batchv1.Job, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.JOB])
	defer txn.Abort()

	cns, cn := client.Namespaced(fqn)
	jj := make([]*batchv1.Job, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		jo, ok := o.(*batchv1.Job)
		if !ok {
			return nil, fmt.Errorf("expecting job")
		}
		if jo.Namespace != cns {
			continue
		}
		for _, o := range jo.OwnerReferences {
			if o.Controller == nil || !*o.Controller {
				continue
			}
			if o.Name == cn {
				jj = append(jj, jo)
			}
		}
	}

	return jj, nil
}

func (db *DB) FindPods(ns string, sel map[string]string) ([]*v1.Pod, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	pp := make([]*v1.Pod, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		po, ok := o.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("expecting pod but got %T", o)
		}
		if po.Namespace != ns {
			continue
		}
		if MatchLabels(po.Labels, sel) {
			pp = append(pp, po)
		}
	}

	return pp, nil
}

func (db *DB) FindPodsBySel(ns string, sel *metav1.LabelSelector) ([]*v1.Pod, error) {
	if sel == nil || sel.Size() == 0 {
		return nil, fmt.Errorf("no pod selector given")
	}

	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.PO])
	defer txn.Abort()
	pp := make([]*v1.Pod, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		po, ok := o.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("expecting pod")
		}
		if po.Namespace != ns {
			continue
		}
		if MatchSelector(po.Labels, sel) {
			pp = append(pp, po)
		}
	}

	return pp, nil
}

func (db *DB) FindNSBySel(sel *metav1.LabelSelector) ([]*v1.Namespace, error) {
	if sel == nil || sel.Size() == 0 {
		return nil, nil
	}

	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.NS])
	defer txn.Abort()
	nss := make([]*v1.Namespace, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		ns, ok := o.(*v1.Namespace)
		if !ok {
			return nil, fmt.Errorf("expecting namespace")
		}
		if MatchSelector(ns.Labels, sel) {
			nss = append(nss, ns)
		}
	}

	return nss, nil
}

func (db *DB) FindNS(ns string) (*v1.Namespace, error) {
	txn := db.Txn(false)
	defer txn.Abort()
	o, err := txn.First(internal.Glossary[internal.NS].String(), "ns", ns)
	if err != nil {
		return nil, err
	}
	nss, ok := o.(*v1.Namespace)
	if !ok {
		return nil, fmt.Errorf("expecting namespace but got %s", o)
	}

	return nss, nil
}

func (db *DB) FindNSNameBySel(sel *metav1.LabelSelector) ([]string, error) {
	if sel == nil || sel.Size() == 0 {
		return nil, nil
	}

	txn := db.Txn(false)
	defer txn.Abort()
	txn, it := db.MustITFor(internal.Glossary[internal.NS])
	defer txn.Abort()
	nss := make([]string, 0, 10)
	for o := it.Next(); o != nil; o = it.Next() {
		ns, ok := o.(*v1.Namespace)
		if !ok {
			return nil, fmt.Errorf("expecting namespace but got %s", o)
		}
		if MatchSelector(ns.Labels, sel) {
			nss = append(nss, ns.Name)
		}
	}

	return nss, nil
}

// Helpers...

// MatchSelector check if pod labels match a selector.
func MatchSelector(labels map[string]string, sel *metav1.LabelSelector) bool {
	if len(labels) == 0 || sel.Size() == 0 {
		return false
	}
	if MatchLabels(labels, sel.MatchLabels) {
		return true
	}

	return matchExp(labels, sel.MatchExpressions)
}

func matchExp(labels map[string]string, ee []metav1.LabelSelectorRequirement) bool {
	for _, e := range ee {
		if matchSel(labels, e) {
			return true
		}
	}

	return false
}

func matchSel(labels map[string]string, e metav1.LabelSelectorRequirement) bool {
	_, ok := labels[e.Key]
	if e.Operator == metav1.LabelSelectorOpDoesNotExist && !ok {
		return true
	}
	if !ok {
		return false
	}

	switch e.Operator {
	case metav1.LabelSelectorOpNotIn:
		for _, v := range e.Values {
			if v1, ok := labels[e.Key]; ok && v1 == v {
				return false
			}
		}
		return true
	case metav1.LabelSelectorOpIn:
		for _, v := range e.Values {
			if v == labels[e.Key] {
				return true
			}
		}
		return false
	case metav1.LabelSelectorOpExists:
		return true
	}

	return false
}

// MatchLabels check if pod labels match a selector.
func MatchLabels(labels, sel map[string]string) bool {
	if len(sel) == 0 {
		return false
	}

	var count int
	for k, v := range sel {
		if v1, ok := labels[k]; ok && v == v1 {
			count++
		}
	}

	return count == len(sel)
}

func (db *DB) Exists(kind types.GVR, fqn string) bool {
	txn := db.Txn(false)
	defer txn.Abort()
	o, err := txn.First(kind.String(), "id", fqn)

	return err == nil && o != nil
}
