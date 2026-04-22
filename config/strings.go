package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Strings accepts either a YAML scalar or a YAML sequence of strings and
// normalizes both to a []string. This lets a when: filter write either
// `os: linux` or `os: [linux, darwin]`.
type Strings []string

// UnmarshalYAML implements scalar-or-sequence parsing.
func (s *Strings) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var v string
		if err := node.Decode(&v); err != nil {
			return err
		}
		if v == "" {
			*s = nil
			return nil
		}
		*s = Strings{v}
		return nil
	case yaml.SequenceNode:
		var v []string
		if err := node.Decode(&v); err != nil {
			return err
		}
		*s = Strings(v)
		return nil
	default:
		return fmt.Errorf("expected string or list of strings, got %v", node.Tag)
	}
}

// MarshalYAML renders a single-element list as a scalar and multi-element lists
// as sequences. This keeps round-trips stable for existing scalar-form configs.
func (s Strings) MarshalYAML() (any, error) {
	switch len(s) {
	case 0:
		return nil, nil
	case 1:
		return s[0], nil
	default:
		return []string(s), nil
	}
}

// Contains reports whether the list contains a case-insensitive match for v.
// An empty list matches anything (callers should short-circuit on len == 0).
func (s Strings) Contains(v string) bool {
	for _, item := range s {
		if strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}

// String joins the list with commas for display purposes.
func (s Strings) String() string {
	return strings.Join(s, ",")
}
