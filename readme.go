package espanso

import (
	"fmt"
	"os"
	"text/template"
)

// Readme represents an espanso readme.
type Readme struct {
	author    string
	longDesc  string
	name      string
	repo      string
	shortDesc string
	title     string
	version   string
}

// NewReadme is generating a new readme.
func NewReadme() Readme {
	return Readme{}
}

// SetAuthor sets the author.
func (r *Readme) SetAuthor(a string) *Readme {
	r.author = a
	return r
}

// SetLongDesc sets the long description. Markdown is supported.
func (r *Readme) SetLongDesc(ld string) *Readme {
	r.longDesc = ld
	return r
}

// SetName sets the name.
func (r *Readme) SetName(n string) *Readme {
	r.name = n
	return r
}

// SetRepo sets the repo.
func (r *Readme) SetRepo(repo string) *Readme {
	r.repo = repo
	return r
}

// SetShortDesc sets the short description.
func (r *Readme) SetShortDesc(sd string) *Readme {
	r.shortDesc = sd
	return r
}

// SetTitle sets the title.
func (r *Readme) SetTitle(t string) *Readme {
	r.title = t
	return r
}

// SetVersion sets the version.
func (r *Readme) SetVersion(v string) *Readme {
	r.version = v
	return r
}

// Author returns the author.
func (r *Readme) Author() string {
	return r.author
}

// LongDesc returns the long description.
func (r *Readme) LongDesc() string {
	return r.longDesc
}

// Name returns the name.
func (r *Readme) Name() string {
	return r.name
}

// Repo returns the repo.
func (r *Readme) Repo() string {
	return r.repo
}

// ShortDesc returns the short description.
func (r *Readme) ShortDesc() string {
	return r.shortDesc
}

// Title returns the title.
func (r *Readme) Title() string {
	return r.title
}

// Version returns the version.
func (r *Readme) Version() string {
	return r.version
}

const readmeTemplate = `---
package_name: "{{ .Name }}"
package_title: "{{ .Title }}"
package_desc: "{{ .ShortDesc }}"
package_version: "{{ .Version }}"
package_author: "{{ .Author }}"
package_repo: "{{ .Repo }}"
---
{{ .LongDesc }}`

// Write is writing the generated espanso package.
func (r *Readme) Write(d string) error {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.Mkdir(d, 0644); err != nil {
			return err
		}
	}
	path := fmt.Sprintf("%s/README.md", d)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	t := template.New("readmeTemplate")
	t, err = t.Parse(readmeTemplate)
	if err != nil {
		return err
	}
	if err := t.Execute(f, r); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}
