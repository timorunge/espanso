// Tests for license constructors and LICENSE file writing.

package espanso

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLicenseConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		license   License
		wantYear  string
		wantOwner string
		wantText  string
	}{
		{
			name:      "BSD 3-Clause",
			license:   BSD3Clause("2019-2026", "Test Author"),
			wantYear:  "2019-2026",
			wantOwner: "Test Author",
			wantText:  "Neither the name of the copyright holder",
		},
		{
			name:      "BSD 2-Clause",
			license:   BSD2Clause("2024", "Someone"),
			wantYear:  "2024",
			wantOwner: "Someone",
			wantText:  "Redistribution and use",
		},
		{
			name:      "MIT",
			license:   MIT("2020-2026", "MIT User"),
			wantYear:  "2020-2026",
			wantOwner: "MIT User",
			wantText:  "MIT License",
		},
		{
			name:      "ISC",
			license:   ISC("2023", "ISC User"),
			wantYear:  "2023",
			wantOwner: "ISC User",
			wantText:  "ISC License",
		},
		{
			name:      "Apache 2.0",
			license:   Apache2("2024", "Apache User"),
			wantYear:  "2024",
			wantOwner: "Apache User",
			wantText:  "Apache License",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(tt.license.Text, tt.wantYear) {
				t.Errorf("license text does not contain year %q", tt.wantYear)
			}
			if !strings.Contains(tt.license.Text, tt.wantOwner) {
				t.Errorf("license text does not contain owner %q", tt.wantOwner)
			}
			if !strings.Contains(tt.license.Text, tt.wantText) {
				t.Errorf("license text does not contain %q", tt.wantText)
			}
		})
	}
}

func TestLicenseYearRange(t *testing.T) {
	t.Parallel()

	l := BSD3Clause("2019-2026", "Timo Runge")
	if !strings.Contains(l.Text, "Copyright (c) 2019-2026 Timo Runge") {
		firstLine, _, _ := strings.Cut(l.Text, "\n")
		t.Errorf("expected year range in copyright line, got:\n%s", firstLine)
	}
}

func TestMPL2(t *testing.T) {
	t.Parallel()

	l := MPL2()
	if !strings.Contains(l.Text, "Mozilla Public License") {
		t.Error("MPL2 license text does not contain expected header")
	}
}

func TestLicenseValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		license License
		wantErr bool
	}{
		{
			name:    "valid license",
			license: License{Text: "MIT License..."},
			wantErr: false,
		},
		{
			name:    "empty text",
			license: License{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.license.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLicenseWriteTo(t *testing.T) {
	t.Parallel()

	l := License{Text: "Some license text."}
	var buf bytes.Buffer
	n, err := l.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n != int64(len("Some license text.")) {
		t.Errorf("WriteTo() n = %d, want %d", n, len("Some license text."))
	}
	if buf.String() != "Some license text." {
		t.Errorf("WriteTo() = %q, want %q", buf.String(), "Some license text.")
	}
}

func TestLicenseWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subdir := filepath.Join(dir, "my-package")

	l := BSD3Clause("2024", "Test Author")
	if err := l.WriteFile(subdir); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(subdir, "LICENSE"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(data), "Test Author") {
		t.Error("LICENSE file does not contain author")
	}
}
