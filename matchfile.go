// Standalone match file support for espanso match/*.yml files.

package espanso

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// MatchFile represents a standalone espanso match YAML file.
// Unlike Package, it has no name, parent, or version fields.
type MatchFile struct {
	Imports    []string `yaml:"imports,omitempty"`
	GlobalVars []Var    `yaml:"global_vars,omitempty"`
	Matches    Matches  `yaml:"matches"`
}

// Validate checks that the match file is well-formed.
func (mf MatchFile) Validate() error {
	var errs []error
	if len(mf.Matches) == 0 {
		errs = append(errs, errors.New("at least one match is required"))
	}
	for i, v := range mf.GlobalVars {
		if err := v.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("var[%d]: %w", i, err))
		}
	}
	if err := mf.Matches.Validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// WriteTo marshals the match file to YAML and writes it to w.
func (mf MatchFile) WriteTo(w io.Writer) (int64, error) {
	if err := mf.Validate(); err != nil {
		return 0, fmt.Errorf("validate match file: %w", err)
	}
	return encodeYAML(w, mf)
}

// ReadFrom populates mf by reading and decoding YAML from r.
func (mf *MatchFile) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("read match file yaml: %w", err)
	}
	if err := yaml.Unmarshal(data, mf); err != nil {
		return 0, fmt.Errorf("unmarshal match file yaml: %w", err)
	}
	return int64(len(data)), nil
}

// WriteFile creates dir/filename and writes the match file YAML.
func (mf MatchFile) WriteFile(dir, filename string) error {
	return writeFile(dir, filename, mf)
}

// ReadMatchFile reads and decodes a match YAML file at path.
func ReadMatchFile(path string) (MatchFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return MatchFile{}, fmt.Errorf("open match file: %w", err)
	}

	var mf MatchFile
	_, readErr := mf.ReadFrom(f)
	closeErr := f.Close()

	if readErr != nil {
		return MatchFile{}, readErr
	}
	if closeErr != nil {
		return MatchFile{}, fmt.Errorf("close match file: %w", closeErr)
	}
	return mf, nil
}
