package override

import (
	"encoding/json"

	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type String struct {
	defaultValue string
	except       []*stringExcept
}

type stringExcept struct {
	Match string `json:"match,omitempty" yaml:"match"`
	Value string `json:"value,omitempty" yaml:"value"`

	match glob.Glob
}

func (se *stringExcept) init() (err error) {
	if se.match, err = glob.Compile(se.Match); err != nil {
		return errors.Wrapf(err, "compiling repo pattern %q", se.match)
	}

	return nil
}

func (s *String) Value(name string) string {
	for _, mv := range s.except {
		if mv.match.Match(name) {
			return mv.Value
		}
	}

	return s.defaultValue
}

func (s *String) MarshalJSON() ([]byte, error) {
	if len(s.except) == 0 {
		return json.Marshal(s.defaultValue)
	}

	return json.Marshal(&struct {
		Default string          `json:"default,omitempty"`
		Except  []*stringExcept `json:"except,omitempty"`
	}{
		Default: s.defaultValue,
		Except:  s.except,
	})
}

func (s *String) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dv string
	if err := unmarshal(&dv); err == nil {
		s.defaultValue = dv
		s.except = []*stringExcept{}
		return nil
	}

	var temp struct {
		Default string          `yaml:"default"`
		Except  []*stringExcept `yaml:"except"`
	}
	if err := unmarshal(&temp); err != nil {
		return err
	}

	s.defaultValue = temp.Default
	s.except = temp.Except

	return s.initExcepts()
}

func (s *String) initExcepts() error {
	var errs *multierror.Error

	for _, se := range s.except {
		if err := se.init(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
