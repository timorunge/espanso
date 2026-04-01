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

func TestReadmeReadFrom(t *testing.T) {
	t.Parallel()

	input := "---\n" +
		"package_name: my-pkg\n" +
		"package_title: My Pkg\n" +
		"package_desc: A description.\n" +
		"package_version: 1.0.0\n" +
		"package_author: Author\n" +
		"package_repo: https://example.com\n" +
		"---\n" +
		"# Body\n"

	var r Readme
	n, err := r.ReadFrom(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if n != int64(len(input)) {
		t.Errorf("ReadFrom() n = %d, want %d", n, len(input))
	}
	if r.Name != "my-pkg" {
		t.Errorf("Name = %q, want %q", r.Name, "my-pkg")
	}
	if r.LongDesc != "# Body\n" {
		t.Errorf("LongDesc = %q, want %q", r.LongDesc, "# Body\n")
	}
}

func TestReadmeReadFromNoBody(t *testing.T) {
	t.Parallel()

	input := "---\n" +
		"package_name: my-pkg\n" +
		"package_title: My Pkg\n" +
		"package_desc: A description.\n" +
		"package_version: 1.0.0\n" +
		"package_author: Author\n" +
		"package_repo: https://example.com\n" +
		"---\n"

	var r Readme
	if _, err := r.ReadFrom(strings.NewReader(input)); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if r.LongDesc != "" {
		t.Errorf("LongDesc = %q, want empty", r.LongDesc)
	}
}

func TestReadmeRoundTrip(t *testing.T) {
	t.Parallel()

	original := Readme{
		Name:      "my-package",
		Title:     "My Package",
		ShortDesc: "A short description.",
		Version:   "1.0.0",
		Author:    "Test Author",
		Repo:      "https://github.com/test/repo",
		LongDesc:  "# My Package\n\nLong description here.\n",
	}

	var buf bytes.Buffer
	if _, err := original.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	var decoded Readme
	if _, err := decoded.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Title != original.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, original.Title)
	}
	if decoded.ShortDesc != original.ShortDesc {
		t.Errorf("ShortDesc = %q, want %q", decoded.ShortDesc, original.ShortDesc)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, original.Version)
	}
	if decoded.Author != original.Author {
		t.Errorf("Author = %q, want %q", decoded.Author, original.Author)
	}
	if decoded.Repo != original.Repo {
		t.Errorf("Repo = %q, want %q", decoded.Repo, original.Repo)
	}
	if decoded.LongDesc != original.LongDesc {
		t.Errorf("LongDesc = %q, want %q", decoded.LongDesc, original.LongDesc)
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

func TestReadReadmeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	original := Readme{
		Name: "my-package", Title: "My Package", ShortDesc: "A short description.",
		Version: "1.0.0", Author: "Test Author", Repo: "https://example.com",
		LongDesc: "# Body\n",
	}
	if err := original.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadReadmeFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("ReadReadmeFile() error = %v", err)
	}
	if got.Name != "my-package" {
		t.Errorf("Name = %q, want %q", got.Name, "my-package")
	}
	if got.LongDesc != "# Body\n" {
		t.Errorf("LongDesc = %q, want %q", got.LongDesc, "# Body\n")
	}
}

func TestReadReadmeFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadReadmeFile("/nonexistent/README.md")
	if err == nil {
		t.Error("ReadReadmeFile() expected error for missing file, got nil")
	}
}
