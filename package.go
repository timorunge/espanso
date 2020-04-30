package espanso

import (
	"fmt"
	"os"
	"text/template"
)

// Package represents an espanso package.
type Package struct {
	name    string
	parent  string
	matches Matches
	version string
}

// NewPackage is generating a new package.
func NewPackage() Package {
	return Package{}
}

// SetName is setting the name value for a package.
func (p *Package) SetName(n string) *Package {
	p.name = n
	return p
}

// SetParent is setting the name value for a package.
func (p *Package) SetParent(parent string) *Package {
	p.parent = parent
	return p
}

// SetMatches is setting the matches for a package.
func (p *Package) SetMatches(m Matches) *Package {
	p.matches = m
	return p
}

// SetVersion is setting the version for a package.
func (p *Package) SetVersion(v string) *Package {
	p.version = v
	return p
}

// Name is returning the package name.
func (p *Package) Name() string {
	return p.name
}

// Parent is returning the package parent.
func (p *Package) Parent() string {
	return p.parent
}

// Matches is returning the package matches.
func (p *Package) Matches() Matches {
	return p.matches
}

// Version is returning the package version.
func (p *Package) Version() string {
	return p.version
}

const packageTemplate = `name: {{ .Name }}
parent: {{ .Parent }}
matches:
{{- range $_, $match := .Matches }}
  - trigger: "{{ $match.Trigger }}"
	{{- if $match.ImagePath }}
    image_path: "{{ $match.ImagePath }}"
    {{- else }}
    replace: "{{ $match.Replace }}"
	{{- if $match.PropagateCase }}
    propagate_case: {{ $match.PropagateCase }}
    {{- end -}}
	{{- if $match.Word }}
    word: {{ $match.Word }}
    {{- end -}}
    {{- end -}}
{{- end -}}`

// Write is writing the generated espanso package.
func (p *Package) Write() error {
	d := fmt.Sprintf("%s/%s", p.Name(), p.Version())
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(fmt.Sprintf("%s/package.yml", d))
	if err != nil {
		return err
	}
	defer f.Close()
	t := template.New("packageTemplate")
	t, err = t.Parse(packageTemplate)
	if err != nil {
		return err
	}
	t.Templates()
	if err := t.Execute(f, p); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}
