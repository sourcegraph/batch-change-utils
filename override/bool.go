package override

import (
	"encoding/json"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type Bool struct {
	rules []*boolRule
}

const allPattern = "*"

type boolRule struct {
	pattern  string
	compiled glob.Glob
	value    bool
}

func FromBool(b bool) Bool {
	return Bool{
		rules: []*boolRule{
			{pattern: allPattern, value: b},
		},
	}
}

func (b *Bool) Is(name string) bool {
	// We want the last match to win, so we'll iterate in reverse order.
	for i := len(b.rules) - 1; i >= 0; i-- {
		if b.rules[i].compiled.Match(name) {
			return b.rules[i].value
		}
	}

	// If nothing matched, we'll treat the value as false.
	return false
}

func (b *Bool) MarshalJSON() ([]byte, error) {
	if len(b.rules) == 0 {
		return json.Marshal(false)
	} else if len(b.rules) == 1 && b.rules[0].pattern == allPattern {
		return json.Marshal(b.rules[0].value)
	}

	rules := []map[string]bool{}
	for _, rule := range b.rules {
		rules = append(rules, map[string]bool{
			rule.pattern: rule.value,
		})
	}
	return json.Marshal(rules)
}

func (b *Bool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var all bool
	if err := unmarshal(&all); err == nil {
		b.rules = []*boolRule{{
			pattern: allPattern,
			value:   all,
		}}
	} else {
		rules := []map[string]bool{}
		if err := unmarshal(&rules); err != nil {
			return err
		}

		b.rules = make([]*boolRule, len(rules))
		for i, rule := range rules {
			if len(rule) != 1 {
				return errors.Errorf("unexpected number of elements in the array at entry %d: %d (must be 1)", i, len(rule))
			}
			for pattern, value := range rule {
				var err error
				b.rules[i], err = newBoolRule(pattern, value)
				if err != nil {
					return errors.Wrapf(err, "building rule for array entry %d", i)
				}
			}
		}
	}

	return nil
}

func newBoolRule(pattern string, value bool) (*boolRule, error) {
	compiled, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &boolRule{
		pattern:  pattern,
		compiled: compiled,
		value:    value,
	}, nil
}
