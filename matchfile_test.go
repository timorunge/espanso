// Tests for standalone match file reading and writing.

package espanso

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchFileValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mf      MatchFile
		wantErr bool
	}{
		{
			name: "valid",
			mf: MatchFile{
				Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
			},
			wantErr: false,
		},
		{
			name:    "empty matches",
			mf:      MatchFile{},
			wantErr: true,
		},
		{
			name: "invalid match propagates",
			mf: MatchFile{
				Matches: Matches{{Triggers: []string{}}},
			},
			wantErr: true,
		},
		{
			name: "invalid global var propagates",
			mf: MatchFile{
				GlobalVars: []Var{{Name: "x"}},
				Matches:    Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.mf.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchFileWriteTo(t *testing.T) {
	t.Parallel()

	mf := MatchFile{
		Matches: Matches{
			{Triggers: []string{":hi"}, Replace: "Hello"},
		},
	}

	var buf bytes.Buffer
	n, err := mf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n == 0 {
		t.Error("WriteTo() wrote 0 bytes")
	}
	if !bytes.Contains(buf.Bytes(), []byte("trigger: :hi")) {
		t.Errorf("output missing trigger\n\ngot:\n%s", buf.String())
	}
}

func TestMatchFileWithImports(t *testing.T) {
	t.Parallel()

	mf := MatchFile{
		Imports: []string{"../common.yml"},
		Matches: Matches{
			{Triggers: []string{":hi"}, Replace: "Hello"},
		},
	}

	var buf bytes.Buffer
	if _, err := mf.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("imports:")) {
		t.Errorf("output missing imports\n\ngot:\n%s", buf.String())
	}
}

func TestMatchFileRoundTrip(t *testing.T) {
	t.Parallel()

	original := MatchFile{
		Imports: []string{"../shared.yml"},
		GlobalVars: []Var{
			{Name: "today", Type: "date", Params: map[string]any{"format": "%Y-%m-%d"}},
		},
		Matches: Matches{
			{Triggers: []string{":hi"}, Replace: "Hello {{today}}"},
		},
	}

	var buf bytes.Buffer
	if _, err := original.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	var decoded MatchFile
	if _, err := decoded.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}

	if len(decoded.Imports) != 1 || decoded.Imports[0] != "../shared.yml" {
		t.Errorf("Imports = %v, want [../shared.yml]", decoded.Imports)
	}
	if len(decoded.GlobalVars) != 1 || decoded.GlobalVars[0].Name != "today" {
		t.Errorf("GlobalVars = %v, want [{today date ...}]", decoded.GlobalVars)
	}
	if len(decoded.Matches) != 1 || decoded.Matches[0].Triggers[0] != ":hi" {
		t.Errorf("Matches[0].Triggers = %v, want [:hi]", decoded.Matches[0].Triggers)
	}
}

func TestMatchFileWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mf := MatchFile{
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}

	if err := mf.WriteFile(dir, "email.yml"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "email.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Contains(data, []byte("trigger: :hi")) {
		t.Errorf("file missing expected content\n\ngot:\n%s", data)
	}
}

func TestReadMatchFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mf := MatchFile{
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}
	if err := mf.WriteFile(dir, "test.yml"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadMatchFile(filepath.Join(dir, "test.yml"))
	if err != nil {
		t.Fatalf("ReadMatchFile() error = %v", err)
	}
	if len(got.Matches) != 1 || got.Matches[0].Replace != "Hello" {
		t.Errorf("ReadMatchFile() matches = %v", got.Matches)
	}
}

func TestReadMatchFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadMatchFile("/nonexistent/test.yml")
	if err == nil {
		t.Error("ReadMatchFile() expected error for missing file, got nil")
	}
}
