// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package schema

import (
	"fmt"

	"github.com/derailed/popeye/internal/client"
	"github.com/hashicorp/go-memdb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type MetaAccessor interface {
	GetNamespace() string
	GetName() string
}

var _ MetaAccessor = (*unstructured.Unstructured)(nil)

type fqnIndexer struct{}

func (fqnIndexer) FromArgs(args ...any) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}

	return []byte(args[0].(string)), nil
}

func (fqnIndexer) FromObject(o any) (bool, []byte, error) {
	m, ok := o.(MetaAccessor)
	if !ok {
		return ok, nil, fmt.Errorf("indexer expected MetaAccessor but got %T", o)
	}

	return true, []byte(client.FQN(m.GetNamespace(), m.GetName())), nil
}

type nsIndexer struct{}

func (nsIndexer) FromArgs(args ...any) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}

	return []byte(args[0].(string)), nil
}

func (nsIndexer) FromObject(o any) (bool, []byte, error) {
	m, ok := o.(MetaAccessor)
	if !ok {
		return ok, nil, fmt.Errorf("indexer expected MetaAccessor but got %T", o)
	}

	return true, []byte(m.GetNamespace()), nil
}

func indexFor(table string) *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: table,
		Indexes: map[string]*memdb.IndexSchema{
			"id": {
				Name:    "id",
				Unique:  true,
				Indexer: &fqnIndexer{},
			},
			"ns": {
				Name:    "ns",
				Indexer: &nsIndexer{},
			},
		},
	}
}
