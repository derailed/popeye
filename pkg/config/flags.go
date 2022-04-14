package config

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// BasicAuth tracks basic authentication.
type BasicAuth struct {
	User     *string
	Password *string
}

// PushGateway tracks gateway representations.
type PushGateway struct {
	Address   *string
	BasicAuth BasicAuth
}

func newPushGateway() *PushGateway {
	return &PushGateway{
		Address:   strPtr(""),
		BasicAuth: BasicAuth{User: strPtr(""), Password: strPtr("")},
	}
}

// Flags represents Popeye CLI flags.
type Flags struct {
	*genericclioptions.ConfigFlags

	LintLevel       *string
	Output          *string
	ClearScreen     *bool
	Save            *bool
	OutputFile      *string
	S3Bucket        *string
	S3Region        *string
	S3Endpoint      *string
	CheckOverAllocs *bool
	AllNamespaces   *bool
	Spinach         *string
	Sections        *[]string
	PushGateway     *PushGateway
	InClusterName   *string
	StandAlone      bool
	ActiveNamespace *string
	ForceExitZero   *bool
	MinScore        *int
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		LintLevel:       strPtr(defaultLintLevel),
		Output:          strPtr("standard"),
		AllNamespaces:   boolPtr(false),
		Save:            boolPtr(false),
		OutputFile:      strPtr(""),
		S3Bucket:        strPtr(""),
		S3Region:        strPtr(""),
		S3Endpoint:      strPtr(""),
		InClusterName:   strPtr(""),
		ClearScreen:     boolPtr(false),
		CheckOverAllocs: boolPtr(false),
		Spinach:         strPtr(""),
		Sections:        &[]string{},
		ConfigFlags:     genericclioptions.NewConfigFlags(false),
		PushGateway:     newPushGateway(),
		ForceExitZero:   boolPtr(false),
		MinScore:        intPtr(0),
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

func intPtr(i int) *int {
	return &i
}
