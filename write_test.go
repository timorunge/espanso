// Tests for the WriteAll convenience function.

package espanso

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAll(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pkg := Package{
		Name:    "test-pkg",
		Parent:  "default",
		Version: "1.0.0",
		Matches: Matches{{Triggers: []string{":hi"}, Replace: "Hello"}},
	}
	r := Readme{
		Name: "test-pkg", Title: "Test", ShortDesc: "A test.",
		Version: "1.0.0", Author: "Author", Repo: "https://example.com",
	}
	l := BSD3Clause("2024", "Author")

	if err := WriteAll(dir, pkg, r, l); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	for _, path := range []string{
		filepath.Join(dir, "test-pkg", "1.0.0", "package.yml"),
		filepath.Join(dir, "test-pkg", "README.md"),
		filepath.Join(dir, "test-pkg", "LICENSE"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", path, err)
		}
	}
}

func TestWriteAllValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pkg  Package
		r    Readme
		l    License
	}{
		{
			name: "invalid package",
			pkg:  Package{},
			r: Readme{
				Name: "x", Title: "x", ShortDesc: "x",
				Version: "x", Author: "x", Repo: "x",
			},
			l: BSD3Clause("2024", "Author"),
		},
		{
			name: "invalid readme",
			pkg: Package{
				Name: "x", Parent: "x", Version: "x",
				Matches: Matches{{Triggers: []string{":x"}, Replace: "x"}},
			},
			r: Readme{},
			l: BSD3Clause("2024", "Author"),
		},
		{
			name: "invalid license",
			pkg: Package{
				Name: "x", Parent: "x", Version: "x",
				Matches: Matches{{Triggers: []string{":x"}, Replace: "x"}},
			},
			r: Readme{
				Name: "x", Title: "x", ShortDesc: "x",
				Version: "x", Author: "x", Repo: "x",
			},
			l: License{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := WriteAll(t.TempDir(), tt.pkg, tt.r, tt.l); err == nil {
				t.Error("WriteAll() expected error, got nil")
			}
		})
	}
}
