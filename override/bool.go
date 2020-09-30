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

func (b Bool) MarshalJSON() ([]byte, error) {
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

func (b *Bool) UnmarshalJSON(data []byte) error {
	var all bool
	if err := json.Unmarshal(data, &all); err == nil {
		b.rules = []*boolRule{{
			pattern: allPattern,
			value:   all,
		}}
		return nil
	}

	var bc boolComplex
	if err := json.Unmarshal(data, &bc); err != nil {
		return err
	}

	return bc.hydrate(b)
}

func (b *Bool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var all bool
	if err := unmarshal(&all); err == nil {
		b.rules = []*boolRule{{
			pattern: allPattern,
			value:   all,
		}}
		return nil
	}

	var bc boolComplex
	if err := unmarshal(&bc); err != nil {
		return err
	}

	return bc.hydrate(b)
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

type boolComplex []map[string]bool

func (bc boolComplex) hydrate(b *Bool) error {
	b.rules = make([]*boolRule, len(bc))
	for i, rule := range bc {
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

	return nil
}

// Define equality methods required for cmp to be able to work its magic.

func (a Bool) Equal(b Bool) bool {
	if len(a.rules) != len(b.rules) {
		return false
	}

	for i := range a.rules {
		if a.rules[i].pattern != b.rules[i].pattern || a.rules[i].value != b.rules[i].value {
			return false
		}
	}

	return true
}

func (a boolRule) Equal(b boolRule) bool {
	return a.pattern == b.pattern && a.value == b.value
}
