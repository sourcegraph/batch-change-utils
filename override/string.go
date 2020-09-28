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

	enc := struct {
		Default string              `json:"default,omitempty"`
		Except  []map[string]string `json:"except,omitempty"`
	}{
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

func (s *String) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dv string
	if err := unmarshal(&dv); err == nil {
		s.defaultValue = dv
		s.rules = nil
		return nil
	}

	var temp struct {
		Default string              `yaml:"default"`
		Except  []map[string]string `yaml:"except"`
	}
	if err := unmarshal(&temp); err != nil {
		return err
	}

	s.defaultValue = temp.Default
	if len(temp.Except) > 0 {
		s.rules = make([]*stringRule, len(temp.Except))
		for i, rule := range temp.Except {
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
	}

	return nil
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
