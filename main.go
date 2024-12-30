// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package main

import (
	"github.com/derailed/popeye/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd.Execute()
}
