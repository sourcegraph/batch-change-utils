package overridable

import "testing"

func TestBoolOrStringIs(t *testing.T) {
	for name, tc := range map[string]struct {
		def        BoolOrString
		input      string
		wantParsed interface{}
	}{
		"wildcard false": {
			def: BoolOrString{
				rules: rules{{pattern: allPattern, value: false}},
			},
			input:      "foo",
			wantParsed: false,
		},
		"wildcard true": {
			def: BoolOrString{
				rules: rules{{pattern: allPattern, value: true}},
			},
			input:      "foo",
			wantParsed: true,
		},
		"wildcard string": {
			def: BoolOrString{
				rules: rules{{pattern: allPattern, value: "draft"}},
			},
			input:      "foo",
			wantParsed: "draft",
		},
		"list exhausted": {
			def: BoolOrString{
				rules: rules{{pattern: "bar*", value: true}},
			},
			input:      "foo",
			wantParsed: false,
		},
		"single match": {
			def: BoolOrString{
				rules: rules{{pattern: "bar*", value: true}},
			},
			input:      "bar",
			wantParsed: true,
		},
		"multiple matches": {
			def: BoolOrString{
				rules: rules{
					{pattern: allPattern, value: true},
					{pattern: "bar*", value: false},
				},
			},
			input:      "bar",
			wantParsed: false,
		},
		"multiple matches string": {
			def: BoolOrString{
				rules: rules{
					{pattern: allPattern, value: true},
					{pattern: "bar*", value: "draft"},
				},
			},
			input:      "bar",
			wantParsed: "draft",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := initBoolOrString(&tc.def); err != nil {
				t.Fatal(err)
			}

			if have := tc.def.Value(tc.input); have != tc.wantParsed {
				t.Errorf("unexpected value: have=%v want=%v", have, tc.wantParsed)
			}
		})
	}
}

// initBoolOrString ensures all rules are compiled.
func initBoolOrString(r *BoolOrString) (err error) {
	for i, rule := range r.rules {
		if rule.compiled == nil {
			r.rules[i], err = newRule(rule.pattern, rule.value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
