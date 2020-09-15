package override

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestStringInvalid(t *testing.T) {
	s := String{except: []*stringExcept{
		{Match: "["},
	}}
	if err := s.initExcepts(); err == nil {
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
				except: []*stringExcept{
					{Match: "bar*", Value: "quux"},
				},
			},
			name: "",
			want: "foo",
		},
		"single match": {
			in: String{
				defaultValue: "foo",
				except: []*stringExcept{
					{Match: "bar*", Value: "quux"},
				},
			},
			name: "barfly",
			want: "quux",
		},
		"multiple matches": {
			in: String{
				defaultValue: "foo",
				except: []*stringExcept{
					{Match: "bar*", Value: "quux"},
					{Match: "b*", Value: "baz"},
				},
			},
			name: "barfly",
			want: "quux",
		},
		"all inputs match; empty name": {
			in: String{
				defaultValue: "foo",
				except: []*stringExcept{
					{Match: "*", Value: "quux"},
				},
			},
			name: "",
			want: "quux",
		},
		"all inputs match; non-empty name": {
			in: String{
				defaultValue: "foo",
				except: []*stringExcept{
					{Match: "*", Value: "quux"},
				},
			},
			name: "foo",
			want: "quux",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// This would normally be done via unmarshalling.
			if err := tc.in.initExcepts(); err != nil {
				t.Fatal(err)
			}

			have := tc.in.Value(tc.name)
			if have != tc.want {
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
				except: []*stringExcept{
					{Match: "foo", Value: "bar"},
				},
			},
			want: `{"default":"foo","except":[{"match":"foo","value":"bar"}]}`,
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
				want: String{except: []*stringExcept{}},
			},
			"non-empty string": {
				in:   `"foo"`,
				want: String{defaultValue: "foo", except: []*stringExcept{}},
			},
			"coercible non-string": {
				in:   `42`,
				want: String{defaultValue: "42", except: []*stringExcept{}},
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
    - match: bar
      value: quux`,
				want: String{defaultValue: "foo", except: []*stringExcept{
					{Match: "bar", Value: "quux"},
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
			"array": `[]`,
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

func (a *String) Equal(b *String) bool {
	return a.defaultValue == b.defaultValue && cmp.Equal(a.except, b.except)
}

func (a *stringExcept) Equal(b *stringExcept) bool {
	return a.Match == b.Match && a.Value == b.Value
}
