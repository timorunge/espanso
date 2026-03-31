// Tests for espanso package.yml writing.

package espanso

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pkg     Package
		wantErr bool
	}{
		{
			name: "valid package",
			pkg: Package{
				Name:    "test-pkg",
				Parent:  "default",
				Version: "1.0.0",
				Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
			},
			wantErr: false,
		},
		{
			name:    "missing all fields",
			pkg:     Package{},
			wantErr: true,
		},
		{
			name: "invalid match",
			pkg: Package{
				Name:    "test-pkg",
				Parent:  "default",
				Version: "1.0.0",
				Matches: Matches{{Triggers: []string{}}},
			},
			wantErr: true,
		},
		{
			name: "invalid variable",
			pkg: Package{
				Name:       "test-pkg",
				Parent:     "default",
				Version:    "1.0.0",
				GlobalVars: []Var{{Name: "x"}},
				Matches:    Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.pkg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackageWriteTo(t *testing.T) {
	t.Parallel()

	pkg := Package{
		Name:    "my-package",
		Parent:  "default",
		Version: "1.0.0",
		Matches: Matches{
			{Triggers: []string{":hello"}, Replace: "Hello World"},
			{Triggers: []string{"alh"}, Replace: "although", PropagateCase: true, Word: true},
		},
	}

	var buf bytes.Buffer
	n, err := pkg.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n == 0 {
		t.Error("WriteTo() wrote 0 bytes")
	}

	got := buf.String()

	// Verify key YAML fields are present.
	for _, want := range []string{"name: my-package", "parent: default", "trigger:", "replace:", "propagate_case: true", "word: true"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("output missing %q\n\ngot:\n%s", want, got)
		}
	}
}

func TestPackageWriteToWithGlobalVars(t *testing.T) {
	t.Parallel()

	pkg := Package{
		Name:    "my-package",
		Parent:  "default",
		Version: "1.0.0",
		GlobalVars: []Var{
			{Name: "today", Type: "date", Params: map[string]any{"format": "%Y-%m-%d"}},
		},
		Matches: Matches{
			{Triggers: []string{":date"}, Replace: "Today is {{today}}"},
		},
	}

	var buf bytes.Buffer
	if _, err := pkg.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{"global_vars:", "name: today", "type: date", "format:", "{{today}}"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("output missing %q\n\ngot:\n%s", want, got)
		}
	}
}

func TestPackageWriteToNoGlobalVarsOmitted(t *testing.T) {
	t.Parallel()

	pkg := Package{
		Name:    "my-package",
		Parent:  "default",
		Version: "1.0.0",
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}

	var buf bytes.Buffer
	if _, err := pkg.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	if bytes.Contains(buf.Bytes(), []byte("global_vars:")) {
		t.Errorf("output should not contain 'global_vars:' when empty\n\ngot:\n%s", buf.String())
	}
}

func TestPackageWriteToValidationError(t *testing.T) {
	t.Parallel()

	pkg := Package{} // Missing required fields.
	var buf bytes.Buffer
	_, err := pkg.WriteTo(&buf)
	if err == nil {
		t.Error("WriteTo() expected validation error, got nil")
	}
}

func TestPackageReadFrom(t *testing.T) {
	t.Parallel()

	input := "name: test-pkg\nparent: default\nmatches:\n  - trigger: :hi\n    replace: Hello\n"
	var p Package
	n, err := p.ReadFrom(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if n != int64(len(input)) {
		t.Errorf("ReadFrom() n = %d, want %d", n, len(input))
	}
	if p.Name != "test-pkg" {
		t.Errorf("Name = %q, want %q", p.Name, "test-pkg")
	}
	if p.Parent != "default" {
		t.Errorf("Parent = %q, want %q", p.Parent, "default")
	}
	if p.Version != "" {
		t.Errorf("Version = %q, want empty (yaml:\"-\")", p.Version)
	}
	if len(p.Matches) != 1 {
		t.Fatalf("Matches len = %d, want 1", len(p.Matches))
	}
	if p.Matches[0].Triggers[0] != ":hi" {
		t.Errorf("Matches[0].Triggers[0] = %q, want %q", p.Matches[0].Triggers[0], ":hi")
	}
}

func TestPackageRoundTrip(t *testing.T) {
	t.Parallel()

	original := Package{
		Name:    "round-trip",
		Parent:  "default",
		Version: "1.0.0",
		Matches: Matches{
			{Triggers: []string{":hi", ":hey"}, Replace: "Hello", Word: true},
		},
	}

	var buf bytes.Buffer
	if _, err := original.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	var decoded Package
	if _, err := decoded.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Parent != original.Parent {
		t.Errorf("Parent = %q, want %q", decoded.Parent, original.Parent)
	}
	if len(decoded.Matches) != len(original.Matches) {
		t.Fatalf("Matches len = %d, want %d", len(decoded.Matches), len(original.Matches))
	}
	if decoded.Matches[0].Triggers[0] != original.Matches[0].Triggers[0] {
		t.Errorf("Triggers[0] = %q, want %q", decoded.Matches[0].Triggers[0], original.Matches[0].Triggers[0])
	}
}

func TestPackageWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subdir := filepath.Join(dir, "test-pkg", "0.1.0")

	pkg := Package{
		Name:    "test-pkg",
		Parent:  "default",
		Version: "0.1.0",
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}

	if err := pkg.WriteFile(subdir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(subdir, "package.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Contains(data, []byte("name: test-pkg")) {
		t.Errorf("package.yml missing expected content\n\ngot:\n%s", data)
	}
}

func TestReadPackageFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkg := Package{
		Name:    "test-pkg",
		Parent:  "default",
		Version: "1.0.0",
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}
	if err := pkg.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadPackageFile(filepath.Join(dir, "package.yml"))
	if err != nil {
		t.Fatalf("ReadPackageFile() error = %v", err)
	}
	if got.Name != "test-pkg" {
		t.Errorf("Name = %q, want %q", got.Name, "test-pkg")
	}
}

func TestReadPackageFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadPackageFile("/nonexistent/package.yml")
	if err == nil {
		t.Error("ReadPackageFile() expected error for missing file, got nil")
	}
}
