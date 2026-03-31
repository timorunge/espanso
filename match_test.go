// Tests for match types, YAML marshaling, and conversion helpers.

package espanso

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMatchValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		match   Match
		wantErr bool
	}{
		{
			name:    "valid replace match",
			match:   Match{Triggers: []string{":hello"}, Replace: "Hello World"},
			wantErr: false,
		},
		{
			name:    "valid image match",
			match:   Match{Triggers: []string{":logo"}, ImagePath: "/path/to/logo.png"},
			wantErr: false,
		},
		{
			name:    "valid multiple triggers",
			match:   Match{Triggers: []string{":hi", ":hey"}, Replace: "Hello"},
			wantErr: false,
		},
		{
			name:    "no triggers",
			match:   Match{Replace: "Hello"},
			wantErr: true,
		},
		{
			name:    "both replace and image_path",
			match:   Match{Triggers: []string{":x"}, Replace: "foo", ImagePath: "/img.png"},
			wantErr: true,
		},
		{
			name:    "neither replace nor image_path",
			match:   Match{Triggers: []string{":x"}},
			wantErr: true,
		},
		{
			name:    "valid regex match",
			match:   Match{Regex: `\bfoo\b`, Replace: "bar"},
			wantErr: false,
		},
		{
			name:    "both triggers and regex",
			match:   Match{Triggers: []string{":x"}, Regex: `\bfoo\b`, Replace: "bar"},
			wantErr: true,
		},
		{
			name:    "neither triggers nor regex",
			match:   Match{Replace: "bar"},
			wantErr: true,
		},
		{
			name: "invalid var in match",
			match: Match{
				Triggers: []string{":x"},
				Replace:  "y",
				Vars:     []Var{{Name: ""}},
			},
			wantErr: true,
		},
		{
			name: "uppercase_style without propagate_case",
			match: Match{
				Triggers:       []string{"alh"},
				Replace:        "although",
				UppercaseStyle: "capitalize_words",
			},
			wantErr: true,
		},
		{
			name: "valid uppercase_style with propagate_case",
			match: Match{
				Triggers:       []string{"alh"},
				Replace:        "although",
				PropagateCase:  true,
				UppercaseStyle: "capitalize_words",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.match.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchMarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		match Match
		want  string
	}{
		{
			name:  "single trigger with replace",
			match: Match{Triggers: []string{":hello"}, Replace: "Hello World"},
			want:  "trigger: :hello\nreplace: Hello World\n",
		},
		{
			name:  "multiple triggers",
			match: Match{Triggers: []string{":hi", ":hey"}, Replace: "Hello"},
			want:  "triggers:\n    - :hi\n    - :hey\nreplace: Hello\n",
		},
		{
			name:  "image path",
			match: Match{Triggers: []string{":logo"}, ImagePath: "/path/to/logo.png"},
			want:  "trigger: :logo\nimage_path: /path/to/logo.png\n",
		},
		{
			name:  "with propagate_case and word",
			match: Match{Triggers: []string{"alh"}, Replace: "although", PropagateCase: true, Word: true},
			want:  "trigger: alh\nreplace: although\npropagate_case: true\nword: true\n",
		},
		{
			name:  "booleans omitted when false",
			match: Match{Triggers: []string{":x"}, Replace: "y"},
			want:  "trigger: :x\nreplace: y\n",
		},
		{
			name:  "newline in replace",
			match: Match{Triggers: []string{":br"}, Replace: "Best Regards,\nJon Snow"},
			want:  "trigger: :br\nreplace: |-\n    Best Regards,\n    Jon Snow\n",
		},
		{
			name:  "regex trigger",
			match: Match{Regex: `\bfoo\b`, Replace: "bar"},
			want:  "regex: \\bfoo\\b\nreplace: bar\n",
		},
		{
			name: "with vars",
			match: Match{
				Triggers: []string{":now"},
				Replace:  "It's {{mytime}}",
				Vars: []Var{
					{Name: "mytime", Type: "date", Params: map[string]any{"format": "%H:%M"}},
				},
			},
			want: "trigger: :now\nreplace: It's {{mytime}}\nvars:\n    - name: mytime\n      type: date\n      params:\n        format: '%H:%M'\n",
		},
		{
			name: "with label",
			match: Match{
				Triggers: []string{":sig"},
				Replace:  "Best regards",
				Label:    "Email signature",
			},
			want: "trigger: :sig\nreplace: Best regards\nlabel: Email signature\n",
		},
		{
			name: "with propagate_case and uppercase_style",
			match: Match{
				Triggers:       []string{"alh"},
				Replace:        "although",
				PropagateCase:  true,
				UppercaseStyle: "capitalize_words",
			},
			want: "trigger: alh\nreplace: although\npropagate_case: true\nuppercase_style: capitalize_words\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := yaml.Marshal(tt.match)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() =\n%s\nwant:\n%s", string(got), tt.want)
			}
		})
	}
}

func TestMatchUnmarshalYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		wantTriggers []string
		wantReplace  string
	}{
		{
			name:         "single trigger",
			input:        "trigger: :hello\nreplace: world\n",
			wantTriggers: []string{":hello"},
			wantReplace:  "world",
		},
		{
			name:         "triggers sequence",
			input:        "triggers:\n  - :hi\n  - :hey\nreplace: Hello\n",
			wantTriggers: []string{":hi", ":hey"},
			wantReplace:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var m Match
			if err := yaml.Unmarshal([]byte(tt.input), &m); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if len(m.Triggers) != len(tt.wantTriggers) {
				t.Fatalf("Triggers len = %d, want %d", len(m.Triggers), len(tt.wantTriggers))
			}
			for i, want := range tt.wantTriggers {
				if m.Triggers[i] != want {
					t.Errorf("Triggers[%d] = %q, want %q", i, m.Triggers[i], want)
				}
			}
			if m.Replace != tt.wantReplace {
				t.Errorf("Replace = %q, want %q", m.Replace, tt.wantReplace)
			}
		})
	}
}

func TestMatchUnmarshalYAMLVars(t *testing.T) {
	t.Parallel()

	input := "trigger: :now\nreplace: \"It's {{mytime}}\"\nvars:\n  - name: mytime\n    type: date\n    params:\n      format: \"%H:%M\"\n"
	var m Match
	if err := yaml.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(m.Vars) != 1 {
		t.Fatalf("Vars len = %d, want 1", len(m.Vars))
	}
	if m.Vars[0].Name != "mytime" {
		t.Errorf("Vars[0].Name = %q, want %q", m.Vars[0].Name, "mytime")
	}
	if m.Vars[0].Type != "date" {
		t.Errorf("Vars[0].Type = %q, want %q", m.Vars[0].Type, "date")
	}
	format, ok := m.Vars[0].Params["format"]
	if !ok {
		t.Fatal("Vars[0].Params missing 'format' key")
	}
	if format != "%H:%M" {
		t.Errorf("Vars[0].Params[format] = %v, want %%H:%%M", format)
	}
}

func TestMatchUnmarshalYAMLRegex(t *testing.T) {
	t.Parallel()

	input := "regex: \\bfoo\\b\nreplace: bar\n"
	var m Match
	if err := yaml.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if m.Regex != `\bfoo\b` {
		t.Errorf("Regex = %q, want %q", m.Regex, `\bfoo\b`)
	}
	if len(m.Triggers) != 0 {
		t.Errorf("Triggers len = %d, want 0", len(m.Triggers))
	}
}

func TestMatchRoundTrip(t *testing.T) {
	t.Parallel()

	original := Match{
		Triggers:       []string{":hi", ":hey"},
		Replace:        "Hello",
		Vars:           []Var{{Name: "v", Type: "echo", Params: map[string]any{"echo": "hi"}}},
		Label:          "Greeting",
		SearchTerms:    []string{"hello", "greet"},
		PropagateCase:  true,
		UppercaseStyle: "capitalize_words",
		Word:           true,
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Match
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(decoded.Triggers) != len(original.Triggers) {
		t.Fatalf("round-trip Triggers len = %d, want %d", len(decoded.Triggers), len(original.Triggers))
	}
	for i := range original.Triggers {
		if decoded.Triggers[i] != original.Triggers[i] {
			t.Errorf("round-trip Triggers[%d] = %q, want %q", i, decoded.Triggers[i], original.Triggers[i])
		}
	}
	if decoded.Replace != original.Replace {
		t.Errorf("round-trip Replace = %q, want %q", decoded.Replace, original.Replace)
	}
	if len(decoded.Vars) != 1 {
		t.Fatalf("round-trip Vars len = %d, want 1", len(decoded.Vars))
	}
	if decoded.Vars[0].Name != "v" {
		t.Errorf("round-trip Vars[0].Name = %q, want %q", decoded.Vars[0].Name, "v")
	}
	if decoded.Label != original.Label {
		t.Errorf("round-trip Label = %q, want %q", decoded.Label, original.Label)
	}
	if len(decoded.SearchTerms) != len(original.SearchTerms) {
		t.Fatalf("round-trip SearchTerms len = %d, want %d", len(decoded.SearchTerms), len(original.SearchTerms))
	}
	for i := range original.SearchTerms {
		if decoded.SearchTerms[i] != original.SearchTerms[i] {
			t.Errorf("round-trip SearchTerms[%d] = %q, want %q", i, decoded.SearchTerms[i], original.SearchTerms[i])
		}
	}
	if decoded.PropagateCase != original.PropagateCase {
		t.Errorf("round-trip PropagateCase = %v, want %v", decoded.PropagateCase, original.PropagateCase)
	}
	if decoded.UppercaseStyle != original.UppercaseStyle {
		t.Errorf("round-trip UppercaseStyle = %q, want %q", decoded.UppercaseStyle, original.UppercaseStyle)
	}
	if decoded.Word != original.Word {
		t.Errorf("round-trip Word = %v, want %v", decoded.Word, original.Word)
	}
}

func TestVarValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		v       Var
		wantErr bool
	}{
		{
			name:    "valid var",
			v:       Var{Name: "today", Type: "date"},
			wantErr: false,
		},
		{
			name:    "missing name",
			v:       Var{Type: "date"},
			wantErr: true,
		},
		{
			name:    "missing type",
			v:       Var{Name: "today"},
			wantErr: true,
		},
		{
			name:    "missing both",
			v:       Var{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.v.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVarWithListParams(t *testing.T) {
	t.Parallel()

	input := "trigger: :fruit\nreplace: \"{{fruit}}\"\nvars:\n  - name: fruit\n    type: random\n    params:\n      choices:\n        - apple\n        - banana\n        - cherry\n"
	var m Match
	if err := yaml.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(m.Vars) != 1 {
		t.Fatalf("Vars len = %d, want 1", len(m.Vars))
	}
	choices, ok := m.Vars[0].Params["choices"]
	if !ok {
		t.Fatal("Params missing 'choices' key")
	}
	list, ok := choices.([]any)
	if !ok {
		t.Fatalf("choices type = %T, want []any", choices)
	}
	if len(list) != 3 {
		t.Fatalf("choices len = %d, want 3", len(list))
	}
	if list[0] != "apple" {
		t.Errorf("choices[0] = %v, want %q", list[0], "apple")
	}
}

func TestMatchesSetWord(t *testing.T) {
	t.Parallel()

	original := Matches{
		{Triggers: []string{"a"}, Replace: "b", Word: false},
		{Triggers: []string{"c"}, Replace: "d", Word: false},
	}

	modified := original.SetWord(true)

	for i, m := range modified {
		if !m.Word {
			t.Errorf("modified[%d].Word = false, want true", i)
		}
	}
	// Original must be unchanged.
	for i, m := range original {
		if m.Word {
			t.Errorf("original[%d].Word = true, want false (mutation detected)", i)
		}
	}
}

func TestMatchesSetPropagateCase(t *testing.T) {
	t.Parallel()

	original := Matches{
		{Triggers: []string{"a"}, Replace: "b"},
		{Triggers: []string{"c"}, Replace: "d"},
	}

	modified := original.SetPropagateCase(true)

	for i, m := range modified {
		if !m.PropagateCase {
			t.Errorf("modified[%d].PropagateCase = false, want true", i)
		}
	}
	for i, m := range original {
		if m.PropagateCase {
			t.Errorf("original[%d].PropagateCase = true, want false (mutation detected)", i)
		}
	}
}

func TestMatchesSetUppercaseStyle(t *testing.T) {
	t.Parallel()

	original := Matches{
		{Triggers: []string{"a"}, Replace: "b"},
		{Triggers: []string{"c"}, Replace: "d"},
	}

	modified := original.SetUppercaseStyle("capitalize_words")

	for i, m := range modified {
		if m.UppercaseStyle != "capitalize_words" {
			t.Errorf("modified[%d].UppercaseStyle = %q, want %q", i, m.UppercaseStyle, "capitalize_words")
		}
	}
	for i, m := range original {
		if original[i].UppercaseStyle != "" {
			t.Errorf("original[%d].UppercaseStyle = %q, want empty (mutation detected)", i, m.UppercaseStyle)
		}
	}
}

func TestMatchesSort(t *testing.T) {
	t.Parallel()

	original := Matches{
		{Triggers: []string{"c"}, Replace: "3"},
		{Triggers: []string{"a"}, Replace: "1"},
		{Triggers: []string{"b"}, Replace: "2"},
	}

	sorted := original.Sort()

	want := []string{"a", "b", "c"}
	for i, m := range sorted {
		if m.Triggers[0] != want[i] {
			t.Errorf("sorted[%d].Triggers[0] = %q, want %q", i, m.Triggers[0], want[i])
		}
	}
	// Original must be unchanged.
	if original[0].Triggers[0] != "c" {
		t.Error("original was mutated")
	}
}

func TestMatchesSortEmpty(t *testing.T) {
	t.Parallel()

	sorted := Matches(nil).Sort()
	if len(sorted) != 0 {
		t.Errorf("Sort() on nil returned len %d, want 0", len(sorted))
	}
}

func TestMatchesDeduplicate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     Matches
		wantLen   int
		wantFirst string
	}{
		{
			name: "no duplicates",
			input: Matches{
				{Triggers: []string{":a"}, Replace: "A"},
				{Triggers: []string{":b"}, Replace: "B"},
			},
			wantLen:   2,
			wantFirst: ":a",
		},
		{
			name: "exact duplicate",
			input: Matches{
				{Triggers: []string{":a"}, Replace: "A"},
				{Triggers: []string{":a"}, Replace: "A2"},
			},
			wantLen:   1,
			wantFirst: ":a",
		},
		{
			name: "partial overlap drops entire match",
			input: Matches{
				{Triggers: []string{":a", ":b"}, Replace: "AB"},
				{Triggers: []string{":b", ":c"}, Replace: "BC"},
			},
			wantLen:   1,
			wantFirst: ":a",
		},
		{
			name:    "empty",
			input:   Matches{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.input.Deduplicate()
			if len(got) != tt.wantLen {
				t.Fatalf("Deduplicate() len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantLen > 0 && got[0].Triggers[0] != tt.wantFirst {
				t.Errorf("first trigger = %q, want %q", got[0].Triggers[0], tt.wantFirst)
			}
		})
	}
}

func TestMatchesDeduplicateOriginalUnmodified(t *testing.T) {
	t.Parallel()

	original := Matches{
		{Triggers: []string{":a"}, Replace: "A"},
		{Triggers: []string{":a"}, Replace: "A2"},
	}
	_ = original.Deduplicate()
	if len(original) != 2 {
		t.Error("original was mutated")
	}
}

func TestDictToMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dict    []string
		wantLen int
		trigger string
		replace string
	}{
		{
			name:    "single pair",
			dict:    []string{"trigger", "replace"},
			wantLen: 1,
			trigger: "trigger",
			replace: "replace",
		},
		{
			name:    "multiple pairs",
			dict:    []string{"a", "b", "c", "d"},
			wantLen: 2,
			trigger: "a",
			replace: "b",
		},
		{
			name:    "empty dict",
			dict:    []string{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches := DictToMatches(tt.dict)
			if len(matches) != tt.wantLen {
				t.Fatalf("DictToMatches() len = %d, want %d", len(matches), tt.wantLen)
			}
			if tt.wantLen > 0 {
				if matches[0].Triggers[0] != tt.trigger {
					t.Errorf("Triggers[0] = %q, want %q", matches[0].Triggers[0], tt.trigger)
				}
				if matches[0].Replace != tt.replace {
					t.Errorf("Replace = %q, want %q", matches[0].Replace, tt.replace)
				}
			}
		})
	}
}

func TestDictToMatchesPanicsOnOdd(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("DictToMatches did not panic on odd-length slice")
		}
	}()
	odd := []string{"a", "b", "c"}
	DictToMatches(odd) //nolint:staticcheck // intentional odd-length to test panic
}
