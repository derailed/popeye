// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package json

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// SpinachSchema describes spinach schema.
const SpinachSchema = "spinach.json"

var (
	//go:embed schemas/spinach.json
	spinachSchema string
)

// Validator tracks schemas validation.
type Validator struct {
	schemas map[string]gojsonschema.JSONLoader
	loader  *gojsonschema.SchemaLoader
}

// NewValidator returns a new instance.
func NewValidator() *Validator {
	v := Validator{
		schemas: map[string]gojsonschema.JSONLoader{
			SpinachSchema: gojsonschema.NewStringLoader(spinachSchema),
		},
	}
	v.register()

	return &v
}

// Init initializes the schemas.
func (v *Validator) register() {
	v.loader = gojsonschema.NewSchemaLoader()
	v.loader.Validate = true
	for k, s := range v.schemas {
		if err := v.loader.AddSchema(k, s); err != nil {
			log.Error().Err(err).Msgf("schema initialization failed: %q", k)
		}
	}
}

// Validate runs document thru given schema validation.
func (v *Validator) Validate(k string, bb []byte) error {
	var m interface{}
	err := yaml.Unmarshal(bb, &m)
	if err != nil {
		return err
	}

	s, ok := v.schemas[k]
	if !ok {
		return fmt.Errorf("no schema found for: %q", k)
	}
	result, err := gojsonschema.Validate(s, gojsonschema.NewGoLoader(m))
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	slices.SortFunc(result.Errors(), func(a, b gojsonschema.ResultError) int {
		return cmp.Compare(a.Description(), b.Description())
	})
	var errs error
	for _, re := range result.Errors() {
		errs = errors.Join(errs, errors.New(re.Description()))
	}

	return errs
}
