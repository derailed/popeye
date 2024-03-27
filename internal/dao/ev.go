// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WarnEvt = "Warning"
)

type EventInfo struct {
	Kind    string
	Reason  string
	Message string
	Count   int64
}

func (e EventInfo) IsIssue() bool {
	return e.Kind == WarnEvt
}

type EventInfos []EventInfo

func (ee EventInfos) Issues() []string {
	if len(ee) == 0 {
		return nil
	}
	ss := make([]string, 0, len(ee))
	for _, e := range ee {
		if e.IsIssue() {
			ss = append(ss, e.Message)
		}
	}

	return ss
}

type Event struct {
	Table
}

func EventsFor(ctx context.Context, gvr types.GVR, level, kind, fqn string) (EventInfos, error) {
	ns, n := client.Namespaced(fqn)
	f, ok := ctx.Value(internal.KeyFactory).(types.Factory)
	if !ok {
		return nil, nil
	}

	ss := make([]string, 0, 2)
	if level != "" {
		ss = append(ss, fmt.Sprintf("type=%s", level))
	}
	if kind != "" {
		ss = append(ss, fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", n, kind))
	}
	ctx = context.WithValue(ctx, internal.KeyFields, strings.Join(ss, ","))

	var t Table
	t.Init(f, types.NewGVR("v1/events"))

	oo, err := t.List(ctx, ns)
	if err != nil {
		return nil, err
	}
	if len(oo) == 0 {
		log.Debug().Msgf("No events found %s: %s", gvr, fqn)
		return nil, nil
	}

	tt := oo[0].(*metav1.Table)
	ee := make(EventInfos, 0, len(tt.Rows))
	for _, r := range tt.Rows {
		ee = append(ee, EventInfo{
			Kind:    r.Cells[1].(string),
			Reason:  r.Cells[2].(string),
			Message: r.Cells[6].(string),
			Count:   r.Cells[8].(int64),
		})
	}

	return ee, nil
}
