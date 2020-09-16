package override

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestBoolInvalid(t *testing.T) {
	t.Run("invalid glob patterns", func(t *testing.T) {
		for name, in := range map[string]Bool{
			"invalid only pattern": {boolOnlyExcept: boolOnlyExcept{
				Only: []string{"["},
			}},
			"invalid except pattern": {boolOnlyExcept: boolOnlyExcept{
				Except: []string{"["},
			}},
		} {
			t.Run(name, func(t *testing.T) {
				if err := in.init(); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("invalid field combinations", func(t *testing.T) {
		b := true

		for name, in := range map[string]Bool{
			"zero value": {},
			"boolean value and only": {
				booleanValue:   &b,
				boolOnlyExcept: boolOnlyExcept{Only: []string{"*"}},
			},
			"boolean value and except": {
				booleanValue:   &b,
				boolOnlyExcept: boolOnlyExcept{Except: []string{"*"}},
			},
			"only and except": {boolOnlyExcept: boolOnlyExcept{
				Only:   []string{"*"},
				Except: []string{"*"},
			}},
			"ALL THE THINGS #yolo": {
				booleanValue: &b,
				boolOnlyExcept: boolOnlyExcept{
					Only:   []string{"*"},
					Except: []string{"*"},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				err := in.init()
				if err == nil {
					t.Error("unexpected nil error")
				} else if _, ok := err.(*boolValidator); !ok {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
			})
		}
	})
}

func TestBoolIs(t *testing.T) {
	bf := false
	bt := true

	for name, tc := range map[string]struct {
		in   Bool
		name string
		want bool
	}{
		"boolean false": {
			in:   Bool{booleanValue: &bf},
			name: "foo",
			want: false,
		},
		"boolean true": {
			in:   Bool{booleanValue: &bt},
			name: "foo",
			want: true,
		},
		"only list; no match": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Only: []string{"bar*"}}},
			name: "foo",
			want: false,
		},
		"only list; with match": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Only: []string{"bar*"}}},
			name: "bar",
			want: true,
		},
		"except list; no match": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Except: []string{"bar*"}}},
			name: "foo",
			want: true,
		},
		"except list; with match": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Except: []string{"bar*"}}},
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
	b := false

	for name, tc := range map[string]struct {
		in   Bool
		want string
	}{
		"boolean value": {
			in:   Bool{booleanValue: &b},
			want: `false`,
		},
		"only list": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Only: []string{"*"}}},
			want: `{"only":["*"]}`,
		},
		"except list": {
			in:   Bool{boolOnlyExcept: boolOnlyExcept{Except: []string{"*"}}},
			want: `{"except":["*"]}`,
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
		b := false

		for name, tc := range map[string]struct {
			in   string
			want Bool
		}{
			"boolean value": {
				in:   `false`,
				want: Bool{booleanValue: &b},
			},
			"only list": {
				in: `only: ["foo", "bar"]`,
				want: Bool{boolOnlyExcept: boolOnlyExcept{
					Only: []string{"foo", "bar"},
				}},
			},
			"except list": {
				in: `except: ["foo", "bar"]`,
				want: Bool{boolOnlyExcept: boolOnlyExcept{
					Except: []string{"foo", "bar"},
				}},
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
			"array":           `[]`,
			"bad only list":   `only: ["["]`,
			"bad except list": `except: ["["]`,
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

func TestBoolValidator(t *testing.T) {
	// It's questionable how useful this test is, but we do technically expose
	// boolValidator as an error value. Let's sanity check the return values of
	// Error(), at least.
	for name, bv := range map[string]boolValidator{
		"not enough fields": {},
		"too many fields":   {"foo", "bar"},
	} {
		t.Run(name, func(t *testing.T) {
			if bv.Error() == "" {
				t.Error("unexpected empty string")
			}
		})
	}

	t.Run("just the right number of fields", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("unexpected lack of panic")
			}
		}()

		bv := boolValidator{"foo"}
		_ = bv.Error()
	})
}

// Define equality methods required for cmp to be able to work its magic.

func (a *Bool) Equal(b *Bool) bool {
	if !cmp.Equal(a.booleanValue, b.booleanValue) {
		return false
	}
	if !cmp.Equal(&a.boolOnlyExcept, &b.boolOnlyExcept) {
		return false
	}
	return true
}

func (a *boolOnlyExcept) Equal(b *boolOnlyExcept) bool {
	if !cmp.Equal(a.Only, b.Only) || !cmp.Equal(a.Except, b.Except) {
		return false
	}
	return true
}
