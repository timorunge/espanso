// Match types and conversion helpers for espanso trigger/replace pairs.

package espanso

import (
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// Match represents a single espanso match rule.
// Either Replace or ImagePath must be set, not both.
// If len(Triggers) == 1, the YAML key is "trigger" (string).
// If len(Triggers) > 1, the YAML key is "triggers" (sequence).
type Match struct {
	Triggers      []string
	Replace       string
	ImagePath     string
	PropagateCase bool
	Word          bool
}

// Validate checks that a match is well-formed.
func (m Match) Validate() error {
	var errs []error
	if len(m.Triggers) == 0 {
		errs = append(errs, errors.New("at least one trigger is required"))
	}
	hasReplace := m.Replace != ""
	hasImage := m.ImagePath != ""
	if hasReplace && hasImage {
		errs = append(errs, errors.New("replace and image_path are mutually exclusive"))
	}
	if !hasReplace && !hasImage {
		errs = append(errs, errors.New("either replace or image_path is required"))
	}
	return errors.Join(errs...)
}

// MarshalYAML produces the espanso match YAML representation.
func (m Match) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	if len(m.Triggers) == 1 {
		appendStringPair(node, "trigger", m.Triggers[0])
	} else {
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, t := range m.Triggers {
			seq.Content = append(seq.Content, scalarNode(t))
		}
		node.Content = append(node.Content, keyNode("triggers"), seq)
	}

	if m.ImagePath != "" {
		appendStringPair(node, "image_path", m.ImagePath)
	} else {
		appendStringPair(node, "replace", m.Replace)
	}

	if m.PropagateCase {
		appendBoolPair(node, "propagate_case", true)
	}
	if m.Word {
		appendBoolPair(node, "word", true)
	}

	return node, nil
}

// Matches is an ordered slice of Match values.
type Matches []Match

// SetWord returns a new Matches with Word set on every element.
func (m Matches) SetWord(w bool) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].Word = w
	}
	return out
}

// SetPropagateCase returns a new Matches with PropagateCase set on every element.
func (m Matches) SetPropagateCase(p bool) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].PropagateCase = p
	}
	return out
}

// DictToMatches converts a flat string slice of alternating trigger/replace
// pairs into Matches. Panics if len(dict) is odd.
func DictToMatches(dict []string) Matches {
	if len(dict)%2 != 0 {
		panic(fmt.Sprintf("espanso: DictToMatches requires even-length slice, got %d", len(dict)))
	}
	matches := make(Matches, 0, len(dict)/2)
	for i := 0; i < len(dict); i += 2 {
		matches = append(matches, Match{
			Triggers: []string{dict[i]},
			Replace:  dict[i+1],
		})
	}
	return matches
}

func keyNode(key string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
}

func scalarNode(val string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val}
}

func appendStringPair(node *yaml.Node, key, val string) {
	node.Content = append(node.Content, keyNode(key), scalarNode(val))
}

func appendBoolPair(node *yaml.Node, key string, val bool) {
	v := "false"
	if val {
		v = "true"
	}
	node.Content = append(node.Content,
		keyNode(key),
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: v},
	)
}
