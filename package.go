// Package writing for espanso package.yml output.

package espanso

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Package represents an espanso package.yml file.
type Package struct {
	Name       string  `yaml:"name"`
	Parent     string  `yaml:"parent"`
	GlobalVars []Var   `yaml:"global_vars,omitempty"`
	Matches    Matches `yaml:"matches"`

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
	for i, v := range p.GlobalVars {
		if err := v.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("variable[%d]: %w", i, err))
		}
	}
	for i := range p.Matches {
		if err := p.Matches[i].Validate(); err != nil {
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

// ReadFrom populates p by reading and decoding YAML from r.
// Version is not populated because it is not stored in package.yml.
func (p *Package) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("read package yaml: %w", err)
	}
	if err := yaml.Unmarshal(data, p); err != nil {
		return 0, fmt.Errorf("unmarshal package yaml: %w", err)
	}
	return int64(len(data)), nil
}

// WriteFile creates dir/package.yml and writes the package YAML.
func (p Package) WriteFile(dir string) error {
	return writeFile(dir, "package.yml", p)
}

// ReadPackageDir walks dir recursively and reads every package.yml file found.
func ReadPackageDir(dir string) ([]Package, error) {
	var packages []Package
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "package.yml" {
			return nil
		}
		p, err := ReadPackageFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		packages = append(packages, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}
	return packages, nil
}

// ReadPackageFile reads and decodes a package.yml file at path.
func ReadPackageFile(path string) (Package, error) {
	f, err := os.Open(path)
	if err != nil {
		return Package{}, fmt.Errorf("open package file: %w", err)
	}

	var p Package
	_, readErr := p.ReadFrom(f)
	closeErr := f.Close()

	if readErr != nil {
		return Package{}, readErr
	}
	if closeErr != nil {
		return Package{}, fmt.Errorf("close package file: %w", closeErr)
	}
	return p, nil
}
