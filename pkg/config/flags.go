// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var outputs = []string{
	"standard",
	"jurassic",
	"yaml",
	"json",
	"html",
	"junit",
	"score",
	"prometheus",
}

// Flags represents Popeye CLI flags.
type Flags struct {
	*genericclioptions.ConfigFlags

	PushGateway     *PushGateway
	S3              *S3Info
	LintLevel       *string
	Output          *string
	ClearScreen     *bool
	Save            *bool
	OutputFile      *string
	CheckOverAllocs *bool
	AllNamespaces   *bool
	Spinach         *string
	Sections        *[]string
	InClusterName   *string
	StandAlone      bool
	ActiveNamespace *string
	ForceExitZero   *bool
	MinScore        *int
	LogLevel        *int
	LogFile         *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		LintLevel:       strPtr(defaultLintLevel),
		Output:          strPtr("standard"),
		AllNamespaces:   boolPtr(false),
		Save:            boolPtr(false),
		OutputFile:      strPtr(""),
		S3:              newS3Info(),
		InClusterName:   strPtr(""),
		ClearScreen:     boolPtr(false),
		CheckOverAllocs: boolPtr(false),
		Spinach:         strPtr(""),
		Sections:        &[]string{},
		ConfigFlags:     genericclioptions.NewConfigFlags(false),
		PushGateway:     newPushGateway(),
		ForceExitZero:   boolPtr(false),
		MinScore:        intPtr(0),
		LogLevel:        intPtr(0),
		LogFile:         strPtr(""),
	}
}

func (f *Flags) Validate() error {
	if !IsBoolSet(f.Save) && IsStrSet(f.OutputFile) {
		return errors.New("'--save' must be used in conjunction with 'output-file'")
	}
	if IsBoolSet(f.Save) && IsStrSet(f.S3.Bucket) {
		return errors.New("'--save' cannot be used in conjunction with 's3-bucket'")
	}

	if !in(outputs, f.Output) {
		return fmt.Errorf("invalid output format. [%s]", strings.Join(outputs, ","))
	}

	if IsStrSet(f.Output) && *f.Output == "prometheus" {
		if f.PushGateway == nil || !IsStrSet(f.PushGateway.URL) {
			return errors.New("you must set --push-gtwy-url when prometheus report is enabled")
		}
	}

	return nil
}

func (f *Flags) IsPersistent() bool {
	return IsBoolSet(f.Save) || IsStrSet(f.OutputFile) || (f.S3 != nil && IsStrSet(f.S3.Bucket))
}

// OutputFormat returns the report output format.
func (f *Flags) OutputFormat() string {
	if f.Output != nil && *f.Output != "" {
		return *f.Output
	}

	return "cool"
}

func (f *Flags) Exhaust() string {
	if f.S3 != nil && IsStrSet(f.S3.Bucket) {
		return *f.S3.Bucket
	}
	if IsStrSet(f.OutputFile) {
		return *f.OutputFile
	}

	return ""
}
