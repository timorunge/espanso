// Readme writing for espanso package README.md output.

package espanso

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

	var fm bytes.Buffer
	if _, err := encodeYAML(&fm, r); err != nil {
		return 0, fmt.Errorf("marshal readme front matter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm.Bytes())
	buf.WriteString("---\n")
	if r.LongDesc != "" {
		buf.WriteString(r.LongDesc)
	}

	n, err := w.Write(buf.Bytes())
	if err != nil {
		return int64(n), fmt.Errorf("write readme: %w", err)
	}
	return int64(n), nil
}

// ReadFrom parses a README.md front-matter document from reader.
// The format is "---\n<YAML>\n---\n<optional body>". The YAML block is
// unmarshaled into the front matter fields and the remaining content is
// stored in LongDesc.
func (r *Readme) ReadFrom(reader io.Reader) (int64, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, fmt.Errorf("read readme: %w", err)
	}
	content := string(data)

	const delim = "---\n"
	if !strings.HasPrefix(content, delim) {
		return 0, errors.New("readme: missing opening front matter delimiter")
	}
	rest := content[len(delim):]
	frontMatter, body, found := strings.Cut(rest, delim)
	if !found {
		return 0, errors.New("readme: missing closing front matter delimiter")
	}

	if err := yaml.Unmarshal([]byte(frontMatter), r); err != nil {
		return 0, fmt.Errorf("unmarshal readme front matter: %w", err)
	}
	r.LongDesc = body
	return int64(len(data)), nil
}

// WriteFile creates dir/README.md and writes the readme content.
func (r Readme) WriteFile(dir string) error {
	return writeFile(dir, "README.md", r)
}

// ReadReadmeFile reads and parses a README.md file at path.
func ReadReadmeFile(path string) (Readme, error) {
	f, err := os.Open(path)
	if err != nil {
		return Readme{}, fmt.Errorf("open readme file: %w", err)
	}

	var r Readme
	_, readErr := r.ReadFrom(f)
	closeErr := f.Close()

	if readErr != nil {
		return Readme{}, readErr
	}
	if closeErr != nil {
		return Readme{}, fmt.Errorf("close readme file: %w", closeErr)
	}
	return r, nil
}
