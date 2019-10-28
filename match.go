package espanso

import (
	"strings"
)

// Match represents a single match for espanso.
type Match struct {
	trigger string
	replace string
	word    bool
}

// NewMatch is generating a new match.
func NewMatch() Match {
	return Match{}
}

// SetTrigger is setting the trigger value for a match.
func (m *Match) SetTrigger(t string) *Match {
	m.trigger = t
	return m
}

// SetReplace is setting the replace value for a match.
func (m *Match) SetReplace(r string) *Match {
	m.replace = r
	return m
}

// SetWord is setting the word value for a match.
func (m *Match) SetWord(w bool) *Match {
	m.word = w
	return m
}

// Trigger returns the trigger value for a match.
func (m *Match) Trigger() string {
	return toRaw(m.trigger)
}

// Replace returns the replace value for a match.
func (m *Match) Replace() string {
	return toRaw(m.replace)
}

// Word returns the word value for a match.
func (m *Match) Word() bool {
	return m.word
}

// Matches represents multiple matches for espanso.
type Matches []Match

// SetWord sets the word value for multiple matches.
func (matches Matches) SetWord(w bool) Matches {
	for _, match := range matches {
		match.SetWord(w)
	}
	return matches
}

// DictToMatches is converting a dict with the format of
// var dict = []string{
// 	"trigger", "replace",
// }
// to Matches.
func DictToMatches(dict []string) Matches {
	var matches Matches
	for i := 0; i < len(dict); i += 2 {
		match := NewMatch()
		match.SetTrigger(dict[i])
		match.SetReplace(dict[i+1])
		matches = append(matches, match)
	}
	return matches
}

// toRaw is generating pseudo "raw string literals".
func toRaw(s string) string {
	if strings.Contains(s, "\n") {
		s = strings.Replace(s, "\n", "\\n", -1)
	}
	if strings.Contains(s, "\"") {
		s = strings.Replace(s, "\"", "\\\"", -1)
	}
	return s
}
