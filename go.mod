module github.com/sourcegraph/batch-change-utils

go 1.15

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.5.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-runewidth v0.0.9
	github.com/nsf/termbox-go v0.0.0-20200418040025-38ba6e5628f1
	github.com/pkg/errors v0.9.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

// See: https://github.com/ghodss/yaml/pull/65
replace github.com/ghodss/yaml => github.com/sourcegraph/yaml v1.0.1-0.20200714132230-56936252f152
