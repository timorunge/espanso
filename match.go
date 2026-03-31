// Match types and conversion helpers for espanso trigger/replace pairs.

package espanso

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Match represents a single espanso match rule.
// Either Replace or ImagePath must be set, not both.
// Either Triggers or Regex must be set, not both.
type Match struct {
	Triggers       []string
	Regex          string
	Replace        string
	ImagePath      string
	Vars           []Var
	Label          string
	SearchTerms    []string
	PropagateCase  bool
	UppercaseStyle string
	Word           bool
}

// Validate checks that a match is well-formed.
func (m Match) Validate() error {
	var errs []error
	hasTriggers := len(m.Triggers) > 0
	hasRegex := m.Regex != ""

	switch {
	case hasTriggers && hasRegex:
		errs = append(errs, errors.New("triggers and regex are mutually exclusive"))
	case !hasTriggers && !hasRegex:
		errs = append(errs, errors.New("either triggers or regex is required"))
	}

	hasReplace := m.Replace != ""
	hasImage := m.ImagePath != ""
	if hasReplace && hasImage {
		errs = append(errs, errors.New("replace and image_path are mutually exclusive"))
	}
	if !hasReplace && !hasImage {
		errs = append(errs, errors.New("either replace or image_path is required"))
	}

	for i, v := range m.Vars {
		if err := v.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("var[%d]: %w", i, err))
		}
	}

	if m.UppercaseStyle != "" && !m.PropagateCase {
		errs = append(errs, errors.New("uppercase_style requires propagate_case"))
	}

	return errors.Join(errs...)
}

// MarshalYAML produces the espanso match YAML representation.
func (m Match) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	switch {
	case m.Regex != "":
		appendStringPair(node, "regex", m.Regex)
	case len(m.Triggers) == 1:
		appendStringPair(node, "trigger", m.Triggers[0])
	default:
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

	if len(m.Vars) > 0 {
		varsNode := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, v := range m.Vars {
			varNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			appendStringPair(varNode, "name", v.Name)
			appendStringPair(varNode, "type", v.Type)
			if len(v.Params) > 0 {
				paramsNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				for pk, pv := range v.Params {
					valNode := &yaml.Node{}
					if err := valNode.Encode(pv); err != nil {
						return nil, fmt.Errorf("encode var param %q: %w", pk, err)
					}
					paramsNode.Content = append(paramsNode.Content, keyNode(pk), valNode)
				}
				varNode.Content = append(varNode.Content, keyNode("params"), paramsNode)
			}
			varsNode.Content = append(varsNode.Content, varNode)
		}
		node.Content = append(node.Content, keyNode("vars"), varsNode)
	}

	if m.Label != "" {
		appendStringPair(node, "label", m.Label)
	}

	if len(m.SearchTerms) > 0 {
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, term := range m.SearchTerms {
			seq.Content = append(seq.Content, scalarNode(term))
		}
		node.Content = append(node.Content, keyNode("search_terms"), seq)
	}

	if m.PropagateCase {
		appendBoolPair(node, "propagate_case", true)
	}
	if m.UppercaseStyle != "" {
		appendStringPair(node, "uppercase_style", m.UppercaseStyle)
	}
	if m.Word {
		appendBoolPair(node, "word", true)
	}

	return node, nil
}

// matchYAML is the intermediate type for unmarshaling YAML into a Match.
type matchYAML struct {
	Trigger        string   `yaml:"trigger"`
	Triggers       []string `yaml:"triggers"`
	Regex          string   `yaml:"regex"`
	Replace        string   `yaml:"replace"`
	ImagePath      string   `yaml:"image_path"`
	Vars           []Var    `yaml:"vars"`
	Label          string   `yaml:"label"`
	SearchTerms    []string `yaml:"search_terms"`
	PropagateCase  bool     `yaml:"propagate_case"`
	UppercaseStyle string   `yaml:"uppercase_style"`
	Word           bool     `yaml:"word"`
}

// UnmarshalYAML populates m from a YAML mapping node, handling both the
// single "trigger" key and the "triggers" sequence key.
func (m *Match) UnmarshalYAML(value *yaml.Node) error {
	var raw matchYAML
	if err := value.Decode(&raw); err != nil {
		return err
	}

	hasTrigger := raw.Trigger != ""
	hasTriggers := len(raw.Triggers) > 0
	hasRegex := raw.Regex != ""

	triggerSources := 0
	if hasTrigger || hasTriggers {
		triggerSources++
	}
	if hasRegex {
		triggerSources++
	}
	if triggerSources > 1 {
		return errors.New("trigger/triggers and regex are mutually exclusive")
	}

	switch {
	case hasTrigger && hasTriggers:
		return errors.New("trigger and triggers are mutually exclusive")
	case hasTrigger:
		m.Triggers = []string{raw.Trigger}
	case hasTriggers:
		m.Triggers = raw.Triggers
	}

	m.Regex = raw.Regex
	m.Replace = raw.Replace
	m.ImagePath = raw.ImagePath
	m.Vars = raw.Vars
	m.Label = raw.Label
	m.SearchTerms = raw.SearchTerms
	m.PropagateCase = raw.PropagateCase
	m.UppercaseStyle = raw.UppercaseStyle
	m.Word = raw.Word
	return nil
}

// Var represents an espanso variable definition.
type Var struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Params map[string]any `yaml:"params,omitempty"`
}

// Validate returns an error if Name or Type is empty.
func (v Var) Validate() error {
	var errs []error
	if v.Name == "" {
		errs = append(errs, errors.New("variable name is required"))
	}
	if v.Type == "" {
		errs = append(errs, errors.New("variable type is required"))
	}
	return errors.Join(errs...)
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

// SetUppercaseStyle returns a new Matches with UppercaseStyle set on every element.
func (m Matches) SetUppercaseStyle(s string) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].UppercaseStyle = s
	}
	return out
}

// Sort returns a new Matches sorted alphabetically by each match's first trigger.
func (m Matches) Sort() Matches {
	out := slices.Clone(m)
	slices.SortFunc(out, func(a, b Match) int {
		ka, kb := "", ""
		if len(a.Triggers) > 0 {
			ka = a.Triggers[0]
		}
		if len(b.Triggers) > 0 {
			kb = b.Triggers[0]
		}
		return strings.Compare(ka, kb)
	})
	return out
}

// Deduplicate returns a new Matches with any match removed whose trigger
// was already seen in an earlier match. First occurrence wins.
func (m Matches) Deduplicate() Matches {
	seen := make(map[string]struct{}, len(m))
	out := make(Matches, 0, len(m))
	for i := range m {
		dup := false
		for _, t := range m[i].Triggers {
			if _, ok := seen[t]; ok {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		for _, t := range m[i].Triggers {
			seen[t] = struct{}{}
		}
		out = append(out, m[i])
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
