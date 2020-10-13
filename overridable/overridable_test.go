package overridable

import (
	"encoding/json"
	"testing"
)

func TestRuleInvalid(t *testing.T) {
	if _, err := newRule("[", true); err == nil {
		t.Error("unexpected nil error")
	}
}

func TestRulesMarshalJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		in   rules
		want string
	}{
		"no rules": {
			in:   rules{},
			want: `[]`,
		},
		"one wildcard rule": {
			in:   rules{{pattern: allPattern, value: true}},
			want: `true`,
		},
		"one non-wildcard rule": {
			in:   rules{{pattern: "bar*", value: true}},
			want: `[{"bar*":true}]`,
		},
		"multiple rules": {
			in: rules{
				{pattern: allPattern, value: true},
				{pattern: "bar*", value: false},
				{pattern: "foo*", value: "draft"},
			},
			want: `[{"*":true},{"bar*":false},{"foo*":"draft"}]`,
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
