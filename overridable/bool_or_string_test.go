package overridable

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

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

func TestBoolOrStringMarshalJSON(t *testing.T) {
	bs := BoolOrString{
		rules{
			{pattern: allPattern, value: true},
			{pattern: "bar*", value: false},
			{pattern: "foo*", value: "draft"},
		},
	}
	data, err := json.Marshal(&bs)
	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if have, want := string(data), `[{"*":true},{"bar*":false},{"foo*":"draft"}]`; have != want {
		t.Errorf("unexpected JSON: have=%q want=%q", have, want)
	}
}

func TestBoolOrStringUnmarshalJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want BoolOrString
		}{
			"single bool": {
				in: `true`,
				want: BoolOrString{
					rules: rules{
						{pattern: allPattern, value: true},
					},
				},
			},
			"single string": {
				in: `"draft"`,
				want: BoolOrString{
					rules: rules{
						{pattern: allPattern, value: "draft"},
					},
				},
			},
			"list": {
				in: `[{"foo*":"bar"}]`,
				want: BoolOrString{
					rules: rules{
						{pattern: "foo*", value: "bar"},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have BoolOrString
				if err := json.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected BoolOrString: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"empty object":    `[{}]`,
			"too many fields": `[{"foo": true,"bar":false}]`,
			"invalid glob":    `[{"[":false}]`,
		} {
			t.Run(name, func(t *testing.T) {
				var have BoolOrString
				if err := json.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestBoolOrStringYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want BoolOrString
		}{
			"single false": {
				in: `false`,
				want: BoolOrString{
					rules: rules{
						{pattern: allPattern, value: false},
					},
				},
			},
			"single true": {
				in: `true`,
				want: BoolOrString{
					rules: rules{
						{pattern: allPattern, value: true},
					},
				},
			},
			"empty list": {
				in: `[]`,
				want: BoolOrString{
					rules: rules{},
				},
			},
			"multiple rule list": {
				in: "- \"*\": true\n- github.com/sourcegraph/*: false\n- github.com/sd9/*: draft",
				want: BoolOrString{
					rules: rules{
						{pattern: allPattern, value: true},
						{pattern: "github.com/sourcegraph/*", value: false},
						{pattern: "github.com/sd9/*", value: "draft"},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have BoolOrString
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected BoolOrString: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"empty object":    `- {}`,
			"too many fields": `- {"foo": true, "bar": false}`,
			"invalid glob":    `- "[": false`,
		} {
			t.Run(name, func(t *testing.T) {
				var have BoolOrString
				if err := yaml.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
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
