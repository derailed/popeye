// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dag

import (
	"errors"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

func TestParseVers(t *testing.T) {
	v, _ := semver.ParseTolerant("1.28")

	uu := map[string]struct {
		info version.Info
		err  error
		ver  semver.Version
	}{
		"empty": {
			err: errors.New(`semver parse failed for "." (""|""): strconv.ParseUint: parsing "": invalid syntax`),
		},
		"happy": {
			info: version.Info{Major: "1", Minor: "28"},
			ver:  v,
		},
		"extras": {
			info: version.Info{Major: "1", Minor: "28+"},
			ver:  v,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			v, err := ParseVersion(&u.info)
			if err != nil {
				assert.Equal(t, u.err.Error(), err.Error())
			} else {
				assert.Equal(t, &u.ver, v)
			}
		})
	}
}

func TestMetaFQN(t *testing.T) {
	uu := []struct {
		m metav1.ObjectMeta
		e string
	}{
		{metav1.ObjectMeta{Namespace: "", Name: "fred"}, "fred"},
		{metav1.ObjectMeta{Namespace: "blee", Name: "fred"}, "blee/fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, metaFQN(u.m))
	}
}
