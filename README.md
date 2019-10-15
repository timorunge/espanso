# espanso

[![Go Report](https://goreportcard.com/badge/github.com/timorunge/espanso)](https://goreportcard.com/report/github.com/timorunge/espanso)
[![Build Status](https://travis-ci.org/timorunge/espanso.svg?branch=master)](https://travis-ci.org/timorunge/espanso)

`espanso` is a Go library without dependencies for creating packages for
[espanso](https://espanso.org), the cross-platform Text Expander.

## Install

```go
import "github.com/timorunge/espanso"
```

## Examples

### General

```go
func main() {
	var (
		matches espanso.Matches
		version = "0.1.0"
	)

	m1 := espanso.NewMatch()
	m1.SetTrigger(":espanso")
	m1.SetReplace("Hi there!")
	matches = append(matches, m1)

	m2 := espanso.NewMatch()
	m2.SetTrigger(":br")
	m2.SetReplace("Best Regards,\nJon Snow")
	matches = append(matches, m2)

	p := espanso.NewPackage()
	p.SetName("my-package")
	p.SetParent("default")
	p.SetMatches(matches)
	p.SetVersion(version)
	if err := p.Write(); err != nil {
		panic(err)
	}

	r := espanso.NewReadme()
	r.SetAuthor("Timo Runge")
	r.SetLongDesc(`Long description for my espanso package. Can be Markdown.`)
	r.SetName(p.Name())
	r.SetRepo("https://github.com/timorunge/espanso")
	r.SetShortDesc("Short description for my espanso package.")
	r.SetTitle("My Package")
	r.SetVersion(version)
	if err := r.Write(p.Name()); err != nil {
		panic(err)
	}
}

```

### DictToMatches

```go
func main() {
	var (
		matches = []string{
			":espanso", "Hi there!",
			":br", "Best Regards,\nJon Snow",
		}
		version = "0.1.0"
	)

	p := espanso.NewPackage()
	p.SetName("my-package")
	p.SetParent("default")
	p.SetMatches(espanso.DictToMatches(matches))
	p.SetVersion(version)
	if err := p.Write(); err != nil {
		panic(err)
	}

	r := espanso.NewReadme()
	r.SetAuthor("Timo Runge")
	r.SetLongDesc(`Long description for my espanso package. Can be Markdown.`)
	r.SetName(p.Name())
	r.SetRepo("https://github.com/timorunge/espanso")
	r.SetShortDesc("Short description for my espanso package.")
	r.SetTitle("My Package")
	r.SetVersion(version)
	if err := r.Write(p.Name()); err != nil {
		panic(err)
	}
}
```

## License

[BSD 3-Clause "New" or "Revised" License](LICENSE)

## Author Information

- Timo Runge
