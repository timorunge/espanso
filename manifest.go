// Manifest writing for espanso hub _manifest.yml files.

package espanso

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents an espanso hub _manifest.yml file.
type Manifest struct {
	Name        string   `yaml:"name"`               // Package identifier (required).
	Title       string   `yaml:"title"`              // Human-readable title (required).
	Description string   `yaml:"description"`        // Short description (required).
	Version     string   `yaml:"version"`            // Semantic version (required).
	Author      string   `yaml:"author"`             // Author name (required).
	Tags        []string `yaml:"tags,omitempty"`     // Discovery tags (optional).
	Homepage    string   `yaml:"homepage,omitempty"` // Project URL (optional).
}

// Validate checks that all required hub fields are set.
func (m Manifest) Validate() error {
	var errs []error
	if m.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if m.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}
	if m.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}
	if m.Version == "" {
		errs = append(errs, errors.New("version is required"))
	}
	if m.Author == "" {
		errs = append(errs, errors.New("author is required"))
	}
	return errors.Join(errs...)
}

// WriteTo marshals the manifest to YAML and writes it to w.
func (m Manifest) WriteTo(w io.Writer) (int64, error) {
	if err := m.Validate(); err != nil {
		return 0, fmt.Errorf("validate manifest: %w", err)
	}
	return encodeYAML(w, m)
}

// ReadFrom populates m by reading and decoding YAML from r.
func (m *Manifest) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("read manifest yaml: %w", err)
	}
	if err := yaml.Unmarshal(data, m); err != nil {
		return 0, fmt.Errorf("unmarshal manifest yaml: %w", err)
	}
	return int64(len(data)), nil
}

// WriteFile creates dir/_manifest.yml and writes the manifest YAML.
func (m Manifest) WriteFile(dir string) error {
	return writeFile(dir, "_manifest.yml", m)
}

// ReadManifestFile reads and decodes a _manifest.yml file at path.
func ReadManifestFile(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open manifest file: %w", err)
	}

	var m Manifest
	_, readErr := m.ReadFrom(f)
	closeErr := f.Close()

	if readErr != nil {
		return Manifest{}, readErr
	}
	if closeErr != nil {
		return Manifest{}, fmt.Errorf("close manifest file: %w", closeErr)
	}
	return m, nil
}
