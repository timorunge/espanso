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

func TestPackageWriteToValidationError(t *testing.T) {
	t.Parallel()

	pkg := Package{} // Missing required fields.
	var buf bytes.Buffer
	_, err := pkg.WriteTo(&buf)
	if err == nil {
		t.Error("WriteTo() expected validation error, got nil")
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
