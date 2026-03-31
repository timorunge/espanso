// Readme writing for espanso package README.md output.

package espanso

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Readme represents a README.md file with YAML front matter.
type Readme struct {
	Name      string `yaml:"package_name"`
	Title     string `yaml:"package_title"`
	ShortDesc string `yaml:"package_desc"`
	Version   string `yaml:"package_version"`
	Author    string `yaml:"package_author"`
	Repo      string `yaml:"package_repo"`

	// LongDesc is the markdown body below the front matter.
	LongDesc string `yaml:"-"`
}

// Validate checks that all front matter fields are set.
func (r Readme) Validate() error {
	var errs []error
	if r.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if r.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}
	if r.ShortDesc == "" {
		errs = append(errs, errors.New("short description is required"))
	}
	if r.Version == "" {
		errs = append(errs, errors.New("version is required"))
	}
	if r.Author == "" {
		errs = append(errs, errors.New("author is required"))
	}
	if r.Repo == "" {
		errs = append(errs, errors.New("repo is required"))
	}
	return errors.Join(errs...)
}

// WriteTo writes the complete README.md content to w.
func (r Readme) WriteTo(w io.Writer) (int64, error) {
	if err := r.Validate(); err != nil {
		return 0, fmt.Errorf("validate readme: %w", err)
	}

	frontMatter, err := yaml.Marshal(r)
	if err != nil {
		return 0, fmt.Errorf("marshal readme front matter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(frontMatter)
	buf.WriteString("---\n")
	if r.LongDesc != "" {
		buf.WriteString(r.LongDesc)
	}

	n, werr := w.Write(buf.Bytes())
	return int64(n), werr
}

// WriteFile creates dir/README.md and writes the readme content.
func (r Readme) WriteFile(dir string) error {
	return writeFile(dir, "README.md", r)
}
