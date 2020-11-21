// Package env provides types to handle step environments in campaign specs.
package env

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// Environment represents an environment used for a campaign step, which may
// require values to be resolved from the outer environment the executor is
// running within.
type Environment struct {
	vars []variable
}

// MarshalJSON marshals the environment to the array form.
func (e Environment) MarshalJSON() ([]byte, error) {
	// Although we allow inputs of both the array and object variants, we'll
	// always marshal to the array variant, since it's semantically a strict
	// superset of the object variant.
	if e.vars == nil {
		return []byte(`[]`), nil
	}
	return json.Marshal(e.vars)
}

// UnmarshalJSON unmarshals an environment from one of the two supported JSON
// forms: an array, or a string→string object.
func (e *Environment) UnmarshalJSON(data []byte) error {
	// data is either an array or object. (Or invalid.) Let's start by trying to
	// unmarshal it as an array.
	if err := json.Unmarshal(data, &e.vars); err == nil {
		return nil
	}

	// It's an object, then. We need to put it into a map, then convert it into
	// an array of variables.
	kv := make(map[string]string)
	if err := json.Unmarshal(data, &kv); err != nil {
		return err
	}

	e.vars = make([]variable, len(kv))
	i := 0
	for k, v := range kv {
		copy := v
		e.vars[i].name = k
		e.vars[i].value = &copy
		i++
	}

	return nil
}

// UnmarshalYAML unmarshals an environment from one of the two supported YAML
// forms: an array, or a string→string object.
func (e *Environment) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// data is either an array or object. (Or invalid.) Let's start by trying to
	// unmarshal it as an array.
	if err := unmarshal(&e.vars); err == nil {
		return nil
	}

	// It's an object, then. As above, we need to convert this via a map.
	kv := make(map[string]string)
	if err := unmarshal(&kv); err != nil {
		return err
	}

	e.vars = make([]variable, len(kv))
	i := 0
	for k, v := range kv {
		copy := v
		e.vars[i].name = k
		e.vars[i].value = &copy
		i++
	}

	return nil
}

// Resolve resolves the environment, using values from the given outer
// environment to fill in environment values as needed. If an environment
// variable doesn't exist in the outer environment, then an empty string will be
// used as the value.
//
// outer must be an array of strings in the form `KEY=VALUE`. Generally
// speaking, this will be the return value from os.Environ().
func (e Environment) Resolve(outer []string) (map[string]string, error) {
	// Convert the given outer environment into a map.
	omap := make(map[string]string, len(outer))
	for _, v := range outer {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("unable to parse environment variable %q", v)
		}
		omap[kv[0]] = kv[1]
	}

	// Now we can iterate over our own environment and fill in the missing
	// values.
	resolved := make(map[string]string, len(e.vars))
	for _, v := range e.vars {
		if v.value == nil {
			// We don't bother checking if v.name exists in omap here because
			// the default behaviour is what we want anyway: we'll get an empty
			// string (since that's the zero value for a string), and that is
			// the desired outcome if the environment variable isn't set.
			resolved[v.name] = omap[v.name]
		} else {
			resolved[v.name] = *v.value
		}
	}

	return resolved, nil
}