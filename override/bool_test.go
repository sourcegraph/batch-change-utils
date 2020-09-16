package override

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestBoolInvalid(t *testing.T) {
	b := Bool{rules: []boolRule{{pattern: "["}}}
	if err := b.init(); err == nil {
		t.Error("unexpected nil error")
	}
}

func TestBoolIs(t *testing.T) {
	for name, tc := range map[string]struct {
		in   Bool
		name string
		want bool
	}{
		"wildcard false": {
			in: Bool{
				rules: []boolRule{{pattern: allPattern, value: false}},
			},
			name: "foo",
			want: false,
		},
		"wildcard true": {
			in: Bool{
				rules: []boolRule{{pattern: allPattern, value: true}},
			},
			name: "foo",
			want: true,
		},
		"list exhausted": {
			in: Bool{
				rules: []boolRule{{pattern: "bar*", value: true}},
			},
			name: "foo",
			want: false,
		},
		"single match": {
			in: Bool{
				rules: []boolRule{{pattern: "bar*", value: true}},
			},
			name: "bar",
			want: true,
		},
		"multiple matches": {
			in: Bool{
				rules: []boolRule{
					{pattern: allPattern, value: true},
					{pattern: "bar*", value: false},
				},
			},
			name: "bar",
			want: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := tc.in.init(); err != nil {
				t.Fatal(err)
			}

			if have := tc.in.Is(tc.name); have != tc.want {
				t.Errorf("unexpected value: have=%v want=%v", have, tc.want)
			}
		})
	}
}

func TestBoolJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   Bool
		want string
	}{
		"no rules": {
			in: Bool{
				rules: []boolRule{},
			},
			want: `false`,
		},
		"one wildcard rule": {
			in: Bool{
				rules: []boolRule{{pattern: allPattern, value: true}},
			},
			want: `true`,
		},
		"one non-wildcard rule": {
			in: Bool{
				rules: []boolRule{{pattern: "bar*", value: true}},
			},
			want: `[{"bar*":true}]`,
		},
		"multiple rules": {
			in: Bool{
				rules: []boolRule{
					{pattern: allPattern, value: true},
					{pattern: "bar*", value: false},
				},
			},
			want: `[{"*":true},{"bar*":false}]`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(&tc.in)
			if err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
			if string(data) != tc.want {
				t.Errorf("unexpected JSON: have=%q want=%q", string(data), tc.want)
			}
		})
	}
}

func TestBoolYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want Bool
		}{
			"single false": {
				in: `false`,
				want: Bool{
					rules: []boolRule{
						{pattern: allPattern, value: false},
					},
				},
			},
			"single true": {
				in: `true`,
				want: Bool{
					rules: []boolRule{
						{pattern: allPattern, value: true},
					},
				},
			},
			"empty list": {
				in: `[]`,
				want: Bool{
					rules: []boolRule{},
				},
			},
			"multiple rule list": {
				in: "- \"*\": true\n- github.com/sourcegraph/*: false",
				want: Bool{
					rules: []boolRule{
						{pattern: allPattern, value: true},
						{pattern: "github.com/sourcegraph/*", value: false},
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected Bool: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"string":          `foo`,
			"empty object":    `- {}`,
			"too many fields": `- {"foo": true, "bar": false}`,
			"invalid glob":    `- "[": false`,
		} {
			t.Run(name, func(t *testing.T) {
				var have Bool
				if err := yaml.Unmarshal([]byte(in), &have); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

// Define equality methods required for cmp to be able to work its magic.

func (a *Bool) Equal(b *Bool) bool {
	return cmp.Equal(a.rules, b.rules)
}

func (a boolRule) Equal(b boolRule) bool {
	return a.pattern == b.pattern && a.value == b.value
}
