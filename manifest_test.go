// Tests for espanso hub _manifest.yml writing.

package espanso

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestManifestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		m       Manifest
		wantErr bool
	}{
		{
			name: "valid with all required fields",
			m: Manifest{
				Name:        "my-package",
				Title:       "My Package",
				Description: "A test package",
				Version:     "0.1.0",
				Author:      "Test Author",
			},
			wantErr: false,
		},
		{
			name:    "missing all fields",
			m:       Manifest{},
			wantErr: true,
		},
		{
			name: "missing name",
			m: Manifest{
				Title:       "My Package",
				Description: "A test package",
				Version:     "0.1.0",
				Author:      "Test Author",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			m: Manifest{
				Name:    "my-package",
				Title:   "My Package",
				Version: "0.1.0",
				Author:  "Test Author",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.m.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManifestWriteTo(t *testing.T) {
	t.Parallel()

	m := Manifest{
		Name:        "my-package",
		Title:       "My Package",
		Description: "A test package",
		Version:     "0.1.0",
		Author:      "Test Author",
		Tags:        []string{"test", "sample"},
		Homepage:    "https://example.com",
	}

	var buf bytes.Buffer
	n, err := m.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n == 0 {
		t.Error("WriteTo() wrote 0 bytes")
	}

	got := buf.String()
	for _, want := range []string{"name: my-package", "title: My Package", "description: A test package", "version: 0.1.0", "author: Test Author", "tags:", "homepage:"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("output missing %q\n\ngot:\n%s", want, got)
		}
	}
}

func TestManifestWriteToNoOptionalFields(t *testing.T) {
	t.Parallel()

	m := Manifest{
		Name:        "my-package",
		Title:       "My Package",
		Description: "A test package",
		Version:     "0.1.0",
		Author:      "Test Author",
	}

	var buf bytes.Buffer
	if _, err := m.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	got := buf.Bytes()
	if bytes.Contains(got, []byte("tags:")) {
		t.Errorf("output should not contain 'tags:' when empty\n\ngot:\n%s", got)
	}
	if bytes.Contains(got, []byte("homepage:")) {
		t.Errorf("output should not contain 'homepage:' when empty\n\ngot:\n%s", got)
	}
}

func TestManifestRoundTrip(t *testing.T) {
	t.Parallel()

	original := Manifest{
		Name:        "my-package",
		Title:       "My Package",
		Description: "A test package",
		Version:     "0.1.0",
		Author:      "Test Author",
		Tags:        []string{"test", "sample"},
		Homepage:    "https://example.com",
	}

	var buf bytes.Buffer
	if _, err := original.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	var decoded Manifest
	if _, err := decoded.ReadFrom(&buf); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Title != original.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, original.Title)
	}
	if decoded.Description != original.Description {
		t.Errorf("Description = %q, want %q", decoded.Description, original.Description)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, original.Version)
	}
	if decoded.Author != original.Author {
		t.Errorf("Author = %q, want %q", decoded.Author, original.Author)
	}
	if len(decoded.Tags) != 2 || decoded.Tags[0] != "test" {
		t.Errorf("Tags = %v, want [test sample]", decoded.Tags)
	}
	if decoded.Homepage != original.Homepage {
		t.Errorf("Homepage = %q, want %q", decoded.Homepage, original.Homepage)
	}
}

func TestManifestWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	m := Manifest{
		Name:        "my-package",
		Title:       "My Package",
		Description: "A test package",
		Version:     "0.1.0",
		Author:      "Test Author",
	}

	if err := m.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "_manifest.yml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !bytes.Contains(data, []byte("name: my-package")) {
		t.Errorf("_manifest.yml missing expected content\n\ngot:\n%s", data)
	}
}

func TestReadManifestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	m := Manifest{
		Name:        "my-package",
		Title:       "My Package",
		Description: "A test package",
		Version:     "0.1.0",
		Author:      "Test Author",
	}
	if err := m.WriteFile(dir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadManifestFile(filepath.Join(dir, "_manifest.yml"))
	if err != nil {
		t.Fatalf("ReadManifestFile() error = %v", err)
	}
	if got.Name != "my-package" {
		t.Errorf("Name = %q, want %q", got.Name, "my-package")
	}
}

func TestReadManifestFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadManifestFile("/nonexistent/_manifest.yml")
	if err == nil {
		t.Error("ReadManifestFile() expected error for missing file, got nil")
	}
}
