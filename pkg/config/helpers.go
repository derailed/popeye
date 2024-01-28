// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import "regexp"

var invalidPathCharsRX = regexp.MustCompile(`[:/]+`)

func in(oo []string, p *string) bool {
	if p == nil {
		return true
	}
	for _, o := range oo {
		if o == *p {
			return true
		}
	}

	return false
}

// SanitizeFileName ensure file spec is valid.
func SanitizeFileName(name string) string {
	return invalidPathCharsRX.ReplaceAllString(name, "-")
}

// IsStrSet checks string option is set.
func IsStrSet(s *string) bool {
	return s != nil && *s != ""
}

// IsBoolSet checks bool option is set
func IsBoolSet(s *bool) bool {
	return s != nil && *s
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
