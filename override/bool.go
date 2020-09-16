package override

import (
	"encoding/json"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type Bool struct {
	booleanValue *bool
	boolOnlyExcept
}

type boolOnlyExcept struct {
	Only   []string `json:"only,omitempty" yaml:"only"`
	Except []string `json:"except,omitempty" yaml:"except"`

	only   []glob.Glob
	except []glob.Glob
}

func (b *Bool) Is(name string) bool {
	if b.booleanValue != nil {
		return *b.booleanValue
	}

	if len(b.only) > 0 {
		for _, g := range b.only {
			if g.Match(name) {
				return true
			}
		}
		return false
	}

	for _, g := range b.except {
		if g.Match(name) {
			return false
		}
	}
	return true
}

func (b *Bool) MarshalJSON() ([]byte, error) {
	if b.booleanValue != nil {
		return json.Marshal(*b.booleanValue)
	}

	return json.Marshal(&b.boolOnlyExcept)
}

func (b *Bool) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var def bool
	if err := unmarshal(&def); err == nil {
		b.booleanValue = &def
		return nil
	}

	b.booleanValue = nil
	if err := unmarshal(&b.boolOnlyExcept); err != nil {
		return err
	}

	return b.init()
}

func (b *Bool) init() error {
	// Only one field should be set.
	bv := boolValidator{}
	if b.booleanValue != nil {
		bv.appendType("boolean value")
	}
	if len(b.Only) > 0 {
		bv.appendType("only list")
	}
	if len(b.Except) > 0 {
		bv.appendType("except list")
	}
	if err := bv.errorOrNil(); err != nil {
		return err
	}

	// Now compile the patterns, if any.
	if len(b.Only) > 0 {
		var err error
		b.only, err = compilePatterns(b.Only)
		return err
	} else if len(b.Except) > 0 {
		var err error
		b.except, err = compilePatterns(b.Except)
		return err
	}

	return nil
}

func compilePatterns(patterns []string) ([]glob.Glob, error) {
	var errs *multierror.Error

	globs := make([]glob.Glob, len(patterns))
	for i, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "compiling repo pattern %q", pattern))
		} else {
			globs[i] = g
		}
	}

	return globs, errs.ErrorOrNil()
}

type boolValidator []string

func (v boolValidator) Error() string {
	if len(v) == 0 {
		return "Bool values must include a boolean value, an only list, or an except list"
	} else if len(v) > 1 {
		return "Bool values must include only one of a boolean value, an only list, or an except list; this value includes: " + strings.Join(v, ", ")
	}

	panic("attempted to call Error() on a boolValidator that doesn't represent an error condition; an errorOrNil() call is missing")
}

func (v *boolValidator) appendType(name string) { *v = append(*v, name) }

func (v *boolValidator) errorOrNil() error {
	if len(*v) == 1 {
		return nil
	}
	return v
}
