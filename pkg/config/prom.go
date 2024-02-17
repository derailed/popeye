// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

// BasicAuth tracks basic authentication.
type BasicAuth struct {
	User     *string
	Password *string
}

func newBasicAuth() BasicAuth {
	return BasicAuth{
		User:     strPtr(""),
		Password: strPtr(""),
	}
}

// PushGateway tracks prometheus gateway representations.
type PushGateway struct {
	URL       *string
	BasicAuth BasicAuth
}

func newPushGateway() *PushGateway {
	return &PushGateway{
		URL:       strPtr(""),
		BasicAuth: newBasicAuth(),
	}
}
