// Package writing for espanso package.yml output.

package espanso

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Package represents an espanso package.yml file.
type Package struct {
	Name    string  `yaml:"name"`
	Parent  string  `yaml:"parent"`
	Matches Matches `yaml:"matches"`

	// Version is used for the directory path only, not written to YAML.
	Version string `yaml:"-"`
}

// Validate checks that all required fields are set and all matches are valid.
func (p Package) Validate() error {
	var errs []error
	if p.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if p.Parent == "" {
		errs = append(errs, errors.New("parent is required"))
	}
	if p.Version == "" {
		errs = append(errs, errors.New("version is required"))
	}
	for i, m := range p.Matches {
		if err := m.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("match[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// WriteTo marshals the package to YAML and writes it to w.
func (p Package) WriteTo(w io.Writer) (int64, error) {
	if err := p.Validate(); err != nil {
		return 0, fmt.Errorf("validate package: %w", err)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(p); err != nil {
		return 0, fmt.Errorf("encode package yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return 0, fmt.Errorf("close yaml encoder: %w", err)
	}

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// WriteFile creates dir/package.yml and writes the package YAML.
func (p Package) WriteFile(dir string) error {
	return writeFile(dir, "package.yml", p)
}
