package config

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Flags represents Popeye CLI flags.
type Flags struct {
	*genericclioptions.ConfigFlags

	LintLevel          *string
	Output             *string
	ClearScreen        *bool
	Save               *bool
	S3Bucket           *string
	CheckOverAllocs    *bool
	AllNamespaces      *bool
	Spinach            *string
	Sections           *[]string
	PushGatewayAddress *string
	InClusterName      *string
	StandAlone         bool
	ActiveNamespace    *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		LintLevel:          strPtr(defaultLintLevel),
		Output:             strPtr("standard"),
		AllNamespaces:      boolPtr(false),
		Save:               boolPtr(false),
		S3Bucket:           strPtr(""),
		InClusterName:      strPtr(""),
		ClearScreen:        boolPtr(false),
		CheckOverAllocs:    boolPtr(false),
		Spinach:            strPtr(""),
		Sections:           &[]string{},
		ConfigFlags:        genericclioptions.NewConfigFlags(false),
		PushGatewayAddress: strPtr(""),
	}
}

// OutputFormat returns the report output format.
func (f *Flags) OutputFormat() string {
	if f.Output != nil && *f.Output != "" {
		return *f.Output
	}

	return "cool"
}

// ----------------------------------------------------------------------------
// Helpers...

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}
