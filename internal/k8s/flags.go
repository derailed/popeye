package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const defaultLintLevel = "ok"

// Flags represents Popeye CLI flags.
type Flags struct {
	*genericclioptions.ConfigFlags

	LintLevel   *string
	Output      *string
	ClearScreen *bool
	Spinach     *string
	Sections    *[]string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	level, clear, format, blank := defaultLintLevel, false, "standard", ""

	return &Flags{
		LintLevel:   &level,
		Output:      &format,
		ClearScreen: &clear,
		Spinach:     &blank,
		Sections:    &[]string{},
		ConfigFlags: genericclioptions.NewConfigFlags(false)}
}

// OutputFormat returns the report output format.
func (f *Flags) OutputFormat() string {
	if f.Output != nil {
		return *f.Output
	}

	return "cool"
}

// ----------------------------------------------------------------------------
// Helpers...

// IsSet checks if a string flag is set.
func IsSet(s *string) bool {
	return s != nil && *s != ""
}
