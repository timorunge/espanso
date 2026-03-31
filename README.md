# espanso

[![CI](https://github.com/timorunge/espanso/actions/workflows/ci.yml/badge.svg)](https://github.com/timorunge/espanso/actions/workflows/ci.yml)
[![Go Report](https://goreportcard.com/badge/github.com/timorunge/espanso)](https://goreportcard.com/report/github.com/timorunge/espanso)
[![Go Version](https://img.shields.io/github/go-mod/go-version/timorunge/espanso)](https://go.dev/)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/timorunge/espanso.svg)](https://pkg.go.dev/github.com/timorunge/espanso)
[![License](https://img.shields.io/github/license/timorunge/espanso)](LICENSE)

`espanso` is a Go library for creating packages for
[espanso](https://espanso.org), the cross-platform text expander.

## Install

```go
import "github.com/timorunge/espanso"
```

## Usage

### Package with matches

```go
p := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    Matches: espanso.Matches{
        {Triggers: []string{":espanso"}, Replace: "Hi there!"},
        {Triggers: []string{":br"}, Replace: "Best Regards,\nJon Snow"},
        {Triggers: []string{"alh"}, Replace: "although", PropagateCase: true, Word: true},
        {Triggers: []string{":cat"}, ImagePath: "/path/to/image.png"},
    },
}
dir := filepath.Join(p.Name, p.Version)
if err := p.WriteFile(dir); err != nil {
    log.Fatal(err)
}
```

### DictToMatches

Convert flat string slices (e.g. from the misspell library) to matches:

```go
p := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    Matches: espanso.DictToMatches([]string{
        ":espanso", "Hi there!",
        ":br", "Best Regards,\nJon Snow",
    }).SetWord(true).SetPropagateCase(true),
}
dir := filepath.Join(p.Name, p.Version)
if err := p.WriteFile(dir); err != nil {
    log.Fatal(err)
}
```

### README and LICENSE

```go
r := espanso.Readme{
    Name:      "my-package",
    Title:     "My Package",
    ShortDesc: "Short description for my espanso package.",
    Version:   "0.1.0",
    Author:    "Timo Runge",
    Repo:      "https://github.com/timorunge/espanso",
    LongDesc:  "Long description. Supports **Markdown**.\n",
}
if err := r.WriteFile("my-package"); err != nil {
    log.Fatal(err)
}

l := espanso.BSD3Clause("2019-2026", "Timo Runge")
if err := l.WriteFile("my-package"); err != nil {
    log.Fatal(err)
}
```

### io.Writer support

All types implement `io.WriterTo` for flexible output:

```go
var buf bytes.Buffer
p.WriteTo(&buf)     // write YAML to buffer
r.WriteTo(os.Stdout) // write README to stdout
```

## Development

```bash
make check    # Run all quality gates (fmt, tidy, vet, lint, test)
make test     # Run tests with race detector
make lint     # Run golangci-lint
make help     # Show all available targets
```

## License

[BSD 3-Clause "New" or "Revised" License](LICENSE)
