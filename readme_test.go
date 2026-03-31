// Tests for espanso README.md writing.

package espanso

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		readme  Readme
		wantErr bool
	}{
		{
			name: "valid readme",
			readme: Readme{
				Name: "test-pkg", Title: "Test", ShortDesc: "A test.",
				Version: "1.0.0", Author: "Author", Repo: "https://example.com",
			},
			wantErr: false,
		},
		{
			name:    "missing all fields",
			readme:  Readme{},
			wantErr: true,
		},
		{
			name:    "missing title only",
			readme:  Readme{Name: "x", ShortDesc: "x", Version: "x", Author: "x", Repo: "x"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.readme.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadmeWriteTo(t *testing.T) {
	t.Parallel()

	r := Readme{
		Name:      "my-package",
		Title:     "My Package",
		ShortDesc: "A short description.",
		Version:   "1.0.0",
		Author:    "Test Author",
		Repo:      "https://github.com/test/repo",
		LongDesc:  "# My Package\n\nLong description here.\n",
	}

	var buf bytes.Buffer
	n, err := r.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n == 0 {
		t.Error("WriteTo() wrote 0 bytes")
	}

	got := buf.String()

	// Must start and end front matter with ---.
	if !strings.HasPrefix(got, "---\n") {
		t.Error("output does not start with ---")
	}

	// Front matter fields must be present.
	for _, want := range []string{
		"package_name: my-package",
		"package_title: My Package",
		"package_desc: A short description.",
		"package_version: 1.0.0",
		"package_author: Test Author",
		"package_repo: https://github.com/test/repo",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n\ngot:\n%s", want, got)
		}
	}

	// Long description must appear after closing ---.
	parts := strings.SplitN(got, "---\n", 3)
	if len(parts) < 3 {
		t.Fatalf("expected 3 parts split by ---, got %d", len(parts))
	}
	if !strings.Contains(parts[2], "Long description here.") {
		t.Errorf("long description not found after front matter\n\ngot:\n%s", got)
	}
}

func TestReadmeWriteToValidationError(t *testing.T) {
	t.Parallel()

	r := Readme{}
	var buf bytes.Buffer
	_, err := r.WriteTo(&buf)
	if err == nil {
		t.Error("WriteTo() expected validation error, got nil")
	}
}

func TestReadmeWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subdir := filepath.Join(dir, "my-package")

	r := Readme{
		Name: "my-package", Title: "My Package", ShortDesc: "Test.",
		Version: "1.0.0", Author: "Author", Repo: "https://example.com",
		LongDesc: "# Body\n",
	}

	if err := r.WriteFile(subdir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(subdir, "README.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "package_name: my-package") {
		t.Errorf("README.md missing expected content\n\ngot:\n%s", data)
	}
}
