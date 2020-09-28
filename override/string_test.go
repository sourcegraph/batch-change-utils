package override

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestStringInvalid(t *testing.T) {
	if _, err := newStringRule("[", "foo"); err == nil {
		t.Error("unexpected nil error")
	}
}

func TestStringValue(t *testing.T) {
	for name, tc := range map[string]struct {
		in   String
		name string
		want string
	}{
		"zero value": {
			in:   String{},
			name: "foo",
			want: "",
		},
		"default value": {
			in:   String{defaultValue: "foo"},
			name: "foo",
			want: "foo",
		},
		"empty name": {
			in:   String{defaultValue: "foo"},
			name: "",
			want: "foo",
		},
		"no match": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "bar*", value: "quux"},
				},
			},
			name: "",
			want: "foo",
		},
		"single match": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "bar*", value: "quux"},
				},
			},
			name: "barfly",
			want: "quux",
		},
		"multiple matches": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "bar*", value: "quux"},
					{pattern: "b*", value: "baz"},
				},
			},
			name: "barfly",
			want: "baz",
		},
		"all inputs match; empty name": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "*", value: "quux"},
				},
			},
			name: "",
			want: "quux",
		},
		"all inputs match; non-empty name": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "*", value: "quux"},
				},
			},
			name: "foo",
			want: "quux",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// This would normally be done via unmarshalling.
			if err := initString(&tc.in); err != nil {
				t.Fatal(err)
			}

			if have := tc.in.Value(tc.name); have != tc.want {
				t.Errorf("unexpected value: have=%q want=%q", have, tc.want)
			}
		})
	}
}

func TestStringJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   String
		want string
	}{
		"zero value": {
			in:   String{},
			want: `""`,
		},
		"default value only": {
			in:   String{defaultValue: "foo"},
			want: `"foo"`,
		},
		"except values": {
			in: String{
				defaultValue: "foo",
				rules: []*stringRule{
					{pattern: "foo", value: "bar"},
				},
			},
			want: `{"default":"foo","except":[{"foo":"bar"}]}`,
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

func TestStringYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   string
			want String
		}{
			"empty string": {
				in:   `""`,
				want: String{rules: nil},
			},
			"non-empty string": {
				in:   `"foo"`,
				want: String{defaultValue: "foo", rules: nil},
			},
			"coercible non-string": {
				in:   `42`,
				want: String{defaultValue: "42", rules: nil},
			},
			"different object": {
				in:   `foo: bar`,
				want: String{},
			},
			"complex value; only default": {
				in:   `default: foo`,
				want: String{defaultValue: "foo"},
			},
			"complex value; default and exceptions": {
				in: `
default: foo
except:
    - bar: quux`,
				want: String{defaultValue: "foo", rules: []*stringRule{
					{pattern: "bar", value: "quux"},
				}},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var have String
				if err := yaml.Unmarshal([]byte(tc.in), &have); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&have, &tc.want); diff != "" {
					t.Errorf("unexpected String: %s", diff)
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for name, in := range map[string]string{
			"array":         `[]`,
			"invalid match": `except: {"match": "["}`,
		} {
			t.Run(name, func(t *testing.T) {
				var have String
				if err := yaml.Unmarshal([]byte(in), &have); err == nil {
					t.Errorf("unexpected nil error: %v", have)
				}
			})
		}
	})
}

// Define equality methods required for cmp to be able to work its magic.

func (a *String) Equal(b *String) bool {
	return a.defaultValue == b.defaultValue && cmp.Equal(a.rules, b.rules)
}

func (a *stringRule) Equal(b *stringRule) bool {
	return a.pattern == b.pattern && a.value == b.value
}

func initString(s *String) (err error) {
	for i, rule := range s.rules {
		if rule.compiled == nil {
			s.rules[i], err = newStringRule(rule.pattern, rule.value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
