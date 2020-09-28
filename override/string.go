package override

import (
	"encoding/json"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type String struct {
	defaultValue string
	rules        []*stringRule
}

type stringRule struct {
	pattern  string
	compiled glob.Glob
	value    string
}

func FromString(s string) String {
	return String{defaultValue: s}
}

func (s *String) Value(name string) string {
	for i := len(s.rules) - 1; i >= 0; i-- {
		if s.rules[i].compiled.Match(name) {
			return s.rules[i].value
		}
	}

	return s.defaultValue
}

func (s *String) MarshalJSON() ([]byte, error) {
	if len(s.rules) == 0 {
		return json.Marshal(s.defaultValue)
	}

	enc := stringComplex{
		Default: s.defaultValue,
		Except:  make([]map[string]string, len(s.rules)),
	}

	for i := 0; i < len(s.rules); i++ {
		enc.Except[i] = map[string]string{
			s.rules[i].pattern: s.rules[i].value,
		}
	}

	return json.Marshal(enc)
}

func (s *String) UnmarshalJSON(data []byte) error {
	var dv string
	if err := json.Unmarshal(data, &dv); err == nil {
		s.defaultValue = dv
		s.rules = nil
		return nil
	}

	var temp stringComplex
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	return temp.hydrate(s)
}

func (s *String) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dv string
	if err := unmarshal(&dv); err == nil {
		s.defaultValue = dv
		s.rules = nil
		return nil
	}

	var temp stringComplex
	if err := unmarshal(&temp); err != nil {
		return err
	}

	return temp.hydrate(s)
}

func newStringRule(pattern, value string) (*stringRule, error) {
	compiled, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &stringRule{
		pattern:  pattern,
		compiled: compiled,
		value:    value,
	}, nil
}

type stringComplex struct {
	Default string              `json:"default,omitempty" yaml:"default"`
	Except  []map[string]string `json:"except,omitempty" yaml:"except"`
}

func (sc *stringComplex) hydrate(s *String) error {
	s.defaultValue = sc.Default

	if len(sc.Except) > 0 {
		s.rules = make([]*stringRule, len(sc.Except))
		for i, rule := range sc.Except {
			if len(rule) != 1 {
				return errors.Errorf("unexpected number of elements in the array entry %d: %d (must be 1)", i, len(rule))
			}
			for pattern, value := range rule {
				var err error
				s.rules[i], err = newStringRule(pattern, value)
				if err != nil {
					return errors.Wrapf(err, "building rule for array entry %d", i)
				}
			}
		}
	} else {
		s.rules = nil
	}

	return nil
}

// Define equality methods required for cmp to be able to work its magic.

func (a String) Equal(b String) bool {
	if a.defaultValue != b.defaultValue {
		return false
	}

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

func (a stringRule) Equal(b stringRule) bool {
	return a.pattern == b.pattern && a.value == b.value
}
