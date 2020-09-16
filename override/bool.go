package override

import (
	"encoding/json"

	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type Bool struct {
	rules []boolRule
}

const allPattern = "*"

type boolRule struct {
	pattern  string
	compiled glob.Glob
	value    bool
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
		b.rules = []boolRule{{
			pattern: allPattern,
			value:   all,
		}}
	} else {
		rules := []map[string]bool{}
		if err := unmarshal(&rules); err != nil {
			return err
		}

		b.rules = []boolRule{}
		for i, rule := range rules {
			if len(rule) != 1 {
				return errors.Errorf("unexpected number of elements in the array at entry %d: %d (must be 1)", i, len(rule))
			}
			for pattern, value := range rule {
				b.rules = append(b.rules, boolRule{
					pattern: pattern,
					value:   value,
				})
			}
		}
	}

	return b.init()
}

func (b *Bool) init() error {
	var errs *multierror.Error

	for i := range b.rules {
		g, err := glob.Compile(b.rules[i].pattern)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "compiling repo pattern %d: %q", i, b.rules[i].pattern))
		} else {
			b.rules[i].compiled = g
		}
	}

	return errs.ErrorOrNil()
}
