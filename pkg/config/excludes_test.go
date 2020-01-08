package config_test

import (
	"testing"

	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestExclusion(t *testing.T) {
	uu := map[string]struct {
		ex   config.Exclusion
		res  string
		code config.ID
		e    bool
	}{
		"empty": {
			ex:   config.Exclusion{Name: "", Codes: []config.ID{}},
			res:  "fred",
			code: 100,
		},
		"plain_match_both": {
			ex:   config.Exclusion{Name: "fred", Codes: []config.ID{100}},
			res:  "fred",
			code: 100,
			e:    true,
		},
		"plain_match_none": {
			ex:   config.Exclusion{Name: "fred", Codes: []config.ID{100}},
			res:  "blee",
			code: 101,
		},
		"plain_match_name": {
			ex:   config.Exclusion{Name: "fred", Codes: []config.ID{100}},
			res:  "fred",
			code: 200,
		},
		"plain_match_code": {
			ex:   config.Exclusion{Name: "fred", Codes: []config.ID{100}},
			res:  "blee",
			code: 100,
		},
		"rx_match_both": {
			ex:   config.Exclusion{Name: "rx:fred", Codes: []config.ID{100}},
			res:  "freddy",
			code: 100,
			e:    true,
		},
		"rx_match_none": {
			ex:   config.Exclusion{Name: "rx:fred", Codes: []config.ID{100}},
			res:  "frued",
			code: 101,
		},
		"rx_match_name": {
			ex:   config.Exclusion{Name: "rx:fred", Codes: []config.ID{100}},
			res:  "freddy",
			code: 200,
		},
		"rx_match_code": {
			ex:   config.Exclusion{Name: "rx:fred", Codes: []config.ID{100}},
			res:  "blee",
			code: 100,
		},
		"rx_match_all_codes": {
			ex:   config.Exclusion{Name: "rx:fred", Codes: []config.ID{}},
			res:  "freddo",
			code: 100,
			e:    true,
		},
		"plain_match_all_codes": {
			ex:   config.Exclusion{Name: "fred", Codes: []config.ID{}},
			res:  "fred",
			code: 100,
			e:    true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ee := config.Excludes{"test": config.Exclusions{u.ex}}
			assert.Equal(t, u.e, ee.ShouldExclude("test", u.res, u.code))
		})
	}
}

func TestExcludes(t *testing.T) {
	uu := map[string]struct {
		excludes config.Excludes
		section  string
		res      string
		code     config.ID
		e        bool
	}{
		"empty": {
			excludes: config.Excludes{},
			section:  "fred",
			res:      "blee",
			code:     100,
		},
		"plain_no_match": {
			excludes: config.Excludes{
				"fred": {
					config.Exclusion{Name: "aa", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "bb", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "cc", Codes: []config.ID{100, 200, 300}},
				},
			},
			section: "fred",
			res:     "blee",
			code:    100,
		},
		"plain_match": {
			excludes: config.Excludes{
				"fred": {
					config.Exclusion{Name: "aa", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "bb", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "cc", Codes: []config.ID{100, 200, 300}},
				},
			},
			section: "fred",
			res:     "aa",
			code:    100,
			e:       true,
		},
		"rx_match": {
			excludes: config.Excludes{
				"fred": {
					config.Exclusion{Name: `rx:\Ablee`, Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "bb", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "cc", Codes: []config.ID{100, 200, 300}},
				},
			},
			section: "fred",
			res:     "bleeblah",
			code:    100,
			e:       true,
		},
		"rx_no_match": {
			excludes: config.Excludes{
				"fred": {
					config.Exclusion{Name: `rx:\Ablee`, Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "bb", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "cc", Codes: []config.ID{100, 200, 300}},
				},
			},
			section: "fred",
			res:     "blahblee",
			code:    100,
		},
		"rx_match_nocode": {
			excludes: config.Excludes{
				"fred": {
					config.Exclusion{Name: "rx:blee", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "bb", Codes: []config.ID{100, 200, 300}},
					config.Exclusion{Name: "cc", Codes: []config.ID{100, 200, 300}},
				},
			},
			section: "fred",
			res:     "bleeblah",
			code:    101,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.excludes.ShouldExclude(u.section, u.res, u.code))
		})
	}
}
