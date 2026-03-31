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
