// Match types and conversion helpers for espanso trigger/replace pairs.

package espanso

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"

	"gopkg.in/yaml.v3"
)

// Match represents a single espanso match rule.
// Exactly one output type must be set: Replace, ImagePath, Markdown, HTML, or Form.
// Either Triggers or Regex must be set, not both.
type Match struct {
	// Trigger input (exactly one required).
	Triggers []string // Text triggers.
	Regex    string   // Regular expression trigger.

	// Output (exactly one required).
	Replace    string                    // Static text replacement.
	ImagePath  string                    // Image file path.
	Markdown   string                    // Markdown-formatted replacement.
	HTML       string                    // Raw HTML replacement.
	Form       string                    // Form layout with {{field}} placeholders.
	FormFields map[string]map[string]any // Per-field form control configuration.

	// Variables.
	Vars []Var

	// UI hints.
	Label       string   // Description shown in espanso search bar.
	SearchTerms []string // Additional keywords for search.

	// Behavior modifiers.
	PropagateCase  bool   // Mirror trigger casing in replacement.
	UppercaseStyle string // Capitalization style: "uppercase", "capitalize", "capitalize_words".
	Word           bool   // Trigger only at word boundaries.
	LeftWord       bool   // Trigger only when left side is a word boundary.
	RightWord      bool   // Trigger only when right side is a word boundary.
	ForceMode      string // Injection backend: "clipboard" or "keys".
	Paragraph      bool   // Prevent extra newlines in markdown output.

	// Context filters.
	FilterClass string // Restrict to window class.
	FilterTitle string // Restrict to window title.
	FilterExec  string // Restrict to executable name.
	FilterOS    string // Restrict to OS: "linux", "macos", "windows".
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

	outputCount := 0
	for _, s := range []string{m.Replace, m.ImagePath, m.Markdown, m.HTML, m.Form} {
		if s != "" {
			outputCount++
		}
	}
	switch {
	case outputCount > 1:
		errs = append(errs, errors.New("replace, image_path, markdown, html, and form are mutually exclusive"))
	case outputCount == 0:
		errs = append(errs, errors.New("one of replace, image_path, markdown, html, or form is required"))
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

	marshalTriggerInput(node, m)

	if err := marshalOutput(node, m); err != nil {
		return nil, err
	}
	if err := marshalVars(node, m.Vars); err != nil {
		return nil, err
	}

	appendOptionalString(node, "label", m.Label)
	appendOptionalStringSeq(node, "search_terms", m.SearchTerms)

	if m.PropagateCase {
		appendBoolPair(node, "propagate_case", true)
	}
	appendOptionalString(node, "uppercase_style", m.UppercaseStyle)
	if m.Word {
		appendBoolPair(node, "word", true)
	}
	if m.LeftWord {
		appendBoolPair(node, "left_word", true)
	}
	if m.RightWord {
		appendBoolPair(node, "right_word", true)
	}
	appendOptionalString(node, "force_mode", m.ForceMode)
	if m.Paragraph {
		appendBoolPair(node, "paragraph", true)
	}
	appendOptionalString(node, "filter_class", m.FilterClass)
	appendOptionalString(node, "filter_title", m.FilterTitle)
	appendOptionalString(node, "filter_exec", m.FilterExec)
	appendOptionalString(node, "filter_os", m.FilterOS)

	return node, nil
}

// matchYAML is the intermediate type for unmarshaling YAML into a Match.
type matchYAML struct {
	Trigger        string                    `yaml:"trigger"`
	Triggers       []string                  `yaml:"triggers"`
	Regex          string                    `yaml:"regex"`
	Replace        string                    `yaml:"replace"`
	ImagePath      string                    `yaml:"image_path"`
	Markdown       string                    `yaml:"markdown"`
	HTML           string                    `yaml:"html"`
	Form           string                    `yaml:"form"`
	FormFields     map[string]map[string]any `yaml:"form_fields"`
	Vars           []Var                     `yaml:"vars"`
	Label          string                    `yaml:"label"`
	SearchTerms    []string                  `yaml:"search_terms"`
	PropagateCase  bool                      `yaml:"propagate_case"`
	UppercaseStyle string                    `yaml:"uppercase_style"`
	Word           bool                      `yaml:"word"`
	LeftWord       bool                      `yaml:"left_word"`
	RightWord      bool                      `yaml:"right_word"`
	ForceMode      string                    `yaml:"force_mode"`
	Paragraph      bool                      `yaml:"paragraph"`
	FilterClass    string                    `yaml:"filter_class"`
	FilterTitle    string                    `yaml:"filter_title"`
	FilterExec     string                    `yaml:"filter_exec"`
	FilterOS       string                    `yaml:"filter_os"`
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
	m.Markdown = raw.Markdown
	m.HTML = raw.HTML
	m.Form = raw.Form
	m.FormFields = raw.FormFields
	m.Vars = raw.Vars
	m.Label = raw.Label
	m.SearchTerms = raw.SearchTerms
	m.PropagateCase = raw.PropagateCase
	m.UppercaseStyle = raw.UppercaseStyle
	m.Word = raw.Word
	m.LeftWord = raw.LeftWord
	m.RightWord = raw.RightWord
	m.ForceMode = raw.ForceMode
	m.Paragraph = raw.Paragraph
	m.FilterClass = raw.FilterClass
	m.FilterTitle = raw.FilterTitle
	m.FilterExec = raw.FilterExec
	m.FilterOS = raw.FilterOS
	return nil
}

// Var represents an espanso variable definition.
type Var struct {
	Name       string         `yaml:"name"`                  // Variable name referenced in {{name}} placeholders.
	Type       string         `yaml:"type"`                  // Extension type: "date", "echo", "random", "choice", "clipboard", "shell", "script".
	Params     map[string]any `yaml:"params,omitempty"`      // Extension-specific parameters.
	InjectVars *bool          `yaml:"inject_vars,omitempty"` // Set to false to suppress variable injection in Params. Nil means unset (espanso default: true).
	DependsOn  []string       `yaml:"depends_on,omitempty"`  // Explicit evaluation order dependencies.
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

// SetLeftWord returns a new Matches with LeftWord set on every element.
func (m Matches) SetLeftWord(lw bool) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].LeftWord = lw
	}
	return out
}

// SetRightWord returns a new Matches with RightWord set on every element.
func (m Matches) SetRightWord(rw bool) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].RightWord = rw
	}
	return out
}

// SetForceMode returns a new Matches with ForceMode set on every element.
func (m Matches) SetForceMode(mode string) Matches {
	out := slices.Clone(m)
	for i := range out {
		out[i].ForceMode = mode
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
		return cmp.Compare(ka, kb)
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

// Validate checks that all matches are well-formed and that no trigger
// appears in more than one match.
func (m Matches) Validate() error {
	var errs []error
	seen := make(map[string]int, len(m))
	for i := range m {
		if err := m[i].Validate(); err != nil {
			errs = append(errs, fmt.Errorf("match[%d]: %w", i, err))
		}
		for _, t := range m[i].Triggers {
			if prev, ok := seen[t]; ok {
				errs = append(errs, fmt.Errorf("duplicate trigger %q in match[%d] and match[%d]", t, prev, i))
			} else {
				seen[t] = i
			}
		}
	}
	return errors.Join(errs...)
}

// Filter returns a new Matches containing only elements where fn returns true.
func (m Matches) Filter(fn func(Match) bool) Matches {
	out := make(Matches, 0, len(m))
	for i := range m {
		if fn(m[i]) {
			out = append(out, m[i])
		}
	}
	return out
}

// Append returns a new Matches concatenating the receiver with all others.
func (m Matches) Append(others ...Matches) Matches {
	all := append([]Matches{m}, others...)
	return slices.Concat(all...)
}

// DictToMatches converts a flat string slice of alternating trigger/replace
// pairs into Matches. Returns an error if len(dict) is odd.
func DictToMatches(dict []string) (Matches, error) {
	if len(dict)%2 != 0 {
		return nil, fmt.Errorf("espanso: DictToMatches requires even-length slice, got %d", len(dict))
	}
	matches := make(Matches, 0, len(dict)/2)
	for i := 0; i < len(dict); i += 2 {
		matches = append(matches, Match{
			Triggers: []string{dict[i]},
			Replace:  dict[i+1],
		})
	}
	return matches, nil
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

func appendOptionalString(node *yaml.Node, key, val string) {
	if val != "" {
		appendStringPair(node, key, val)
	}
}

func appendOptionalStringSeq(node *yaml.Node, key string, vals []string) {
	if len(vals) == 0 {
		return
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, v := range vals {
		seq.Content = append(seq.Content, scalarNode(v))
	}
	node.Content = append(node.Content, keyNode(key), seq)
}

func marshalTriggerInput(node *yaml.Node, m Match) {
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
}

func marshalOutput(node *yaml.Node, m Match) error {
	switch {
	case m.Form != "":
		appendStringPair(node, "form", m.Form)
		if len(m.FormFields) > 0 {
			fieldsNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			for _, fk := range sortedKeys(m.FormFields) {
				innerNode := &yaml.Node{}
				if err := innerNode.Encode(m.FormFields[fk]); err != nil {
					return fmt.Errorf("encode form field %q: %w", fk, err)
				}
				fieldsNode.Content = append(fieldsNode.Content, keyNode(fk), innerNode)
			}
			node.Content = append(node.Content, keyNode("form_fields"), fieldsNode)
		}
	case m.Markdown != "":
		appendStringPair(node, "markdown", m.Markdown)
	case m.HTML != "":
		appendStringPair(node, "html", m.HTML)
	case m.ImagePath != "":
		appendStringPair(node, "image_path", m.ImagePath)
	default:
		appendStringPair(node, "replace", m.Replace)
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}

func marshalVars(node *yaml.Node, vars []Var) error {
	if len(vars) == 0 {
		return nil
	}
	varsNode := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, v := range vars {
		varNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendStringPair(varNode, "name", v.Name)
		appendStringPair(varNode, "type", v.Type)
		if len(v.Params) > 0 {
			paramsNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			for _, pk := range sortedKeys(v.Params) {
				valNode := &yaml.Node{}
				if err := valNode.Encode(v.Params[pk]); err != nil {
					return fmt.Errorf("encode var param %q: %w", pk, err)
				}
				paramsNode.Content = append(paramsNode.Content, keyNode(pk), valNode)
			}
			varNode.Content = append(varNode.Content, keyNode("params"), paramsNode)
		}
		if v.InjectVars != nil {
			appendBoolPair(varNode, "inject_vars", *v.InjectVars)
		}
		appendOptionalStringSeq(varNode, "depends_on", v.DependsOn)
		varsNode.Content = append(varsNode.Content, varNode)
	}
	node.Content = append(node.Content, keyNode("vars"), varsNode)
	return nil
}
