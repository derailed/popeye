package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNsLinter(t *testing.T) {
	mks := NewMockClient()
	m.When(mks.ListAllNS()).ThenReturn(map[string]v1.Namespace{
		"ns1": makeNS("ns1", true),
		"ns2": makeNS("ns2", false),
	}, nil)

	l := NewNamespace(mks, nil)
	l.Lint(context.Background())

	mks.VerifyWasCalled(pegomock.Times(1)).ListNS()
	assert.Equal(t, 2, len(l.Issues()))
	assert.Equal(t, 1, len(l.Issues()["ns1"]))
	assert.Equal(t, 1, len(l.Issues()["ns2"]))
}

func TestNsLint(t *testing.T) {
	uu := []struct {
		nn     map[string]v1.Namespace
		issues int
	}{
		{
			map[string]v1.Namespace{
				"ns1": makeNS("ns1", true),
				"ns2": makeNS("ns2", true),
			},
			0,
		},
		{
			map[string]v1.Namespace{
				"ns1": makeNS("ns1", true),
				"ns2": makeNS("ns2", false),
			},
			1,
		},
	}

	for _, u := range uu {
		l := NewNamespace(nil, nil)
		l.lint(u.nn, nil)
		assert.Equal(t, len(u.nn), len(l.Issues()))
		var tissue int
		for _, ns := range u.nn {
			tissue += len(l.Issues()[ns.Name])
		}
		assert.Equal(t, u.issues, tissue)
	}
}

func TestNsCheckActive(t *testing.T) {
	uu := []struct {
		active bool
		issues int
	}{
		{true, 0},
		{false, 1},
	}

	for _, u := range uu {
		ns := makeNS("ns1", u.active)
		l := NewNamespace(nil, nil)
		l.checkActive(ns)

		assert.Equal(t, u.issues, len(l.Issues()))
	}
}

func TestNsCheckInUse(t *testing.T) {
	uu := []struct {
		name   string
		issues int
	}{
		{"ns1", 0},
		{"ns2", 1},
	}

	for _, u := range uu {
		ns := makeNS(u.name, true)
		l := NewNamespace(nil, nil)
		l.checkInUse(ns.Name, []string{"ns1"})

		assert.Equal(t, u.issues, len(l.Issues()))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeNS(n string, active bool) v1.Namespace {
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
	}

	ns.Status.Phase = v1.NamespaceTerminating
	if active {
		ns.Status.Phase = v1.NamespaceActive
	}

	return ns
}
