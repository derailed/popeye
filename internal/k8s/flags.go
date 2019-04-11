package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const defaultLintLevel = "ok"

// Flags represents Popeye CLI flags.
type Flags struct {
	*genericclioptions.ConfigFlags

	LintLevel   *string
	Jurassic    *bool
	ClearScreen *bool
	Spinach     *string
	Sections    *[]string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	level, enable, blank := defaultLintLevel, false, ""

	return &Flags{
		LintLevel:   &level,
		ClearScreen: &enable,
		Jurassic:    &enable,
		Spinach:     &blank,
		Sections:    &[]string{},
		ConfigFlags: genericclioptions.NewConfigFlags(false)}
}

// ----------------------------------------------------------------------------
// Helpers...

// IsSet checks if a string flag is set.
func IsSet(s *string) bool {
	return s != nil && *s != ""
}
