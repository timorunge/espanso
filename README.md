# espanso

[![CI](https://github.com/timorunge/espanso/actions/workflows/ci.yml/badge.svg)](https://github.com/timorunge/espanso/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/timorunge/espanso)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/timorunge/espanso)](https://goreportcard.com/report/github.com/timorunge/espanso)
[![License](https://img.shields.io/github/license/timorunge/espanso)](LICENSE)
[![Release](https://img.shields.io/github/v/release/timorunge/espanso)](https://github.com/timorunge/espanso/releases)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/timorunge/espanso.svg)](https://pkg.go.dev/github.com/timorunge/espanso)

`espanso` is a Go library for creating
[espanso](https://espanso.org) packages and match files programmatically.

Compatible with **espanso v2** (tested against v2.2+). Covers the full
match specification: all trigger types, output types, behavior modifiers,
context filters, variables, and form fields.

## Usage

### Write a package

```go
package main

import (
    "log"
    "path/filepath"

    "github.com/timorunge/espanso"
)

func main() {
    p := espanso.Package{
        Name:    "my-package",
        Parent:  "default",
        Version: "0.1.0",
        Matches: espanso.Matches{
            {Triggers: []string{":espanso"}, Replace: "Hi there!"},
            {Triggers: []string{":br"}, Replace: "Best Regards,\nJon Snow"},
            {
                Triggers:       []string{"alh"},
                Replace:        "although",
                PropagateCase:  true,
                UppercaseStyle: "capitalize_words",
                Word:           true,
            },
            {Triggers: []string{":cat"}, ImagePath: "/path/to/image.png"},
        },
    }
    dir := filepath.Join(p.Name, p.Version)
    if err := p.WriteFile(dir); err != nil {
        log.Fatal(err)
    }
}
```

### Variables

Global variables are shared across all matches in a package. Match
variables are scoped to a single match:

```go
p := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    GlobalVars: []espanso.Var{
        {Name: "today", Type: "date", Params: map[string]any{"format": "%Y-%m-%d"}},
    },
    Matches: espanso.Matches{
        {Triggers: []string{":date"}, Replace: "Today is {{today}}"},
        {
            Triggers: []string{":now"},
            Replace:  "It's {{mytime}}",
            Label:    "Current time",
            Vars: []espanso.Var{
                {Name: "mytime", Type: "date", Params: map[string]any{"format": "%H:%M"}},
            },
        },
    },
}
if err := p.WriteFile(filepath.Join(p.Name, p.Version)); err != nil {
    log.Fatal(err)
}
```

### Forms, filters, and output types

```go
// Form with a text input and a dropdown.
form := espanso.Match{
    Triggers: []string{":greet"},
    Form:     "Hello {{name}}, welcome to {{team}}!",
    FormFields: map[string]map[string]any{
        "name": {"type": "text"},
        "team": {"type": "list", "values": []string{"backend", "frontend", "infra"}},
    },
}

// Restrict to a specific application and OS.
filtered := espanso.Match{
    Triggers:    []string{":debug"},
    Replace:     "console.log('debug:', $|$)",
    FilterClass: "Code",
    FilterOS:    "linux",
}

// Markdown output with paragraph mode.
markdown := espanso.Match{
    Triggers:  []string{":sig"},
    Markdown:  "**Best regards**,\n*Jon Snow*",
    Paragraph: true,
}

// HTML output.
html := espanso.Match{
    Triggers: []string{":badge"},
    HTML:     "<span style=\"color:green\">OK</span>",
}

// Regex trigger.
redact := espanso.Match{Regex: `\bemail@\S+`, Replace: "[redacted]"}

p := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    Matches: espanso.Matches{form, filtered, markdown, html, redact},
}
if err := p.WriteFile(filepath.Join(p.Name, p.Version)); err != nil {
    log.Fatal(err)
}
```

### DictToMatches

Convert flat string slices (e.g. from the misspell library) to matches:

```go
matches, err := espanso.DictToMatches([]string{
    ":espanso", "Hi there!",
    ":br", "Best Regards,\nJon Snow",
})
if err != nil {
    log.Fatal(err)
}
p := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    Matches: matches.SetWord(true).SetPropagateCase(true),
}
if err := p.WriteFile(filepath.Join(p.Name, p.Version)); err != nil {
    log.Fatal(err)
}
```

### Filter, Append, Sort, and Deduplicate

```go
matches := espanso.Matches{
    {Triggers: []string{":a"}, Replace: "A"},
    {Regex: `\bfoo\b`, Replace: "bar"},
    {Triggers: []string{":b"}, Replace: "B"},
    {Triggers: []string{":a"}, Replace: "A duplicate"},
}
extra := espanso.Matches{
    {Triggers: []string{":c"}, Replace: "C"},
}

triggers := matches.Filter(func(m espanso.Match) bool {
    return len(m.Triggers) > 0 // drop regex-only matches
})
all := triggers.Append(extra).Sort().Deduplicate()
if err := all.Validate(); err != nil {
    log.Fatal(err)
}
```

### Writing metadata

README and LICENSE files for an espanso package:

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

l := espanso.MIT("2019-2026", "Timo Runge")
if err := l.WriteFile("my-package"); err != nil {
    log.Fatal(err)
}
```

Several license templates are available (MIT, BSD-2-Clause, BSD-3-Clause,
ISC, Apache-2.0, MPL-2.0, CC-BY-SA-3.0, CC-BY-SA-4.0). See
[pkg.go.dev](https://pkg.go.dev/github.com/timorunge/espanso) for the
full list.

`WriteAll` writes package.yml, README.md, and LICENSE in one call:

```go
if err := espanso.WriteAll("output", p, r, l); err != nil {
    log.Fatal(err)
}
// Creates: output/{name}/{version}/package.yml, output/{name}/README.md, output/{name}/LICENSE
```

### Reading packages

```go
p, err := espanso.ReadPackageFile("misspell-en/0.1.2/package.yml")
if err != nil {
    log.Fatal(err)
}
fmt.Println(p.Name) // "misspell-en"
```

Read all packages in a directory tree:

```go
packages, err := espanso.ReadPackageDir(context.Background(), "packages")
if err != nil {
    log.Fatal(err)
}
for _, pkg := range packages {
    fmt.Println(pkg.Name)
}
```

### Standalone match files

Write individual match files (for `match/*.yml`) without full package
structure:

```go
mf := espanso.MatchFile{
    Imports: []string{"../common.yml"},
    Matches: espanso.Matches{
        {Triggers: []string{":sig"}, Replace: "Best Regards,\nJon Snow"},
    },
}
if err := mf.WriteFile("match", "email.yml"); err != nil {
    log.Fatal(err)
}

// Read back.
mf, err := espanso.ReadMatchFile("match/email.yml")
if err != nil {
    log.Fatal(err)
}
fmt.Println(mf.Matches[0].Replace)
```

### Hub registry packages

Create packages compliant with the
[espanso hub](https://hub.espanso.org) registry structure
(`name/version/_manifest.yml` + `package.yml` + `README.md`):

```go
manifest := espanso.Manifest{
    Name:        "my-package",
    Title:       "My Package",
    Description: "Useful text expansions.",
    Version:     "0.1.0",
    Author:      "Jon Snow",
    Tags:        []string{"productivity", "email"},
    Homepage:    "https://github.com/user/my-package",
}
pkg := espanso.Package{
    Name:    "my-package",
    Parent:  "default",
    Version: "0.1.0",
    Matches: espanso.Matches{
        {Triggers: []string{":hi"}, Replace: "Hello!"},
    },
}
r := espanso.Readme{
    Name: "my-package", Title: "My Package", ShortDesc: "Useful text expansions.",
    Version: "0.1.0", Author: "Jon Snow", Repo: "https://github.com/user/my-package",
}

if err := espanso.WriteHubPackage("output", manifest, pkg, r); err != nil {
    log.Fatal(err)
}
// Creates: output/my-package/0.1.0/{_manifest.yml,package.yml,README.md}
```

### io.WriterTo

All writable types implement `io.WriterTo` for flexible output:

```go
var buf bytes.Buffer
mf := espanso.MatchFile{
    Matches: espanso.Matches{
        {Triggers: []string{":hi"}, Replace: "Hello!"},
    },
}
if _, err := mf.WriteTo(&buf); err != nil {
    log.Fatal(err)
}
fmt.Print(buf.String())
```

## Match Field Reference

All fields from the [espanso v2 match specification](https://espanso.org/docs/matches/basics/):

| Field | Type | Description |
|-------|------|-------------|
| `Triggers` | `[]string` | Text triggers (mutually exclusive with `Regex`) |
| `Regex` | `string` | Regex trigger (mutually exclusive with `Triggers`) |
| `Replace` | `string` | Static text replacement |
| `ImagePath` | `string` | Image file path |
| `Markdown` | `string` | Markdown-formatted replacement |
| `HTML` | `string` | Raw HTML replacement |
| `Form` | `string` | Form layout with `{{field}}` placeholders |
| `FormFields` | `map[string]map[string]any` | Per-field form control config |
| `Vars` | `[]Var` | Variable definitions (date, shell, random, etc.) |
| `Label` | `string` | Description shown in search bar |
| `SearchTerms` | `[]string` | Extra search keywords |
| `PropagateCase` | `bool` | Mirror trigger casing in replacement |
| `UppercaseStyle` | `string` | Multi-word capitalization style |
| `Word` | `bool` | Trigger only at word boundaries |
| `LeftWord` | `bool` | Trigger only when left side is a word separator |
| `RightWord` | `bool` | Trigger only when right side is a word separator |
| `ForceMode` | `string` | Injection backend: `clipboard` or `keys` |
| `Paragraph` | `bool` | Prevent extra newlines in markdown output |
| `FilterClass` | `string` | Restrict to window class |
| `FilterTitle` | `string` | Restrict to window title |
| `FilterExec` | `string` | Restrict to executable name |
| `FilterOS` | `string` | Restrict to OS (`linux`, `macos`, `windows`) |

### Var fields

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Variable name used in `{{name}}` placeholders |
| `Type` | `string` | Extension type: `date`, `echo`, `random`, `choice`, `clipboard`, `shell`, `script` |
| `Params` | `map[string]any` | Extension-specific parameters |
| `InjectVars` | `*bool` | Set `false` to suppress variable injection in params |
| `DependsOn` | `[]string` | Explicit evaluation order dependencies |

### Matches collection methods

| Method | Description |
|--------|-------------|
| `SetWord(bool)` | Set `Word` on all matches |
| `SetLeftWord(bool)` | Set `LeftWord` on all matches |
| `SetRightWord(bool)` | Set `RightWord` on all matches |
| `SetPropagateCase(bool)` | Set `PropagateCase` on all matches |
| `SetUppercaseStyle(string)` | Set `UppercaseStyle` on all matches |
| `SetForceMode(string)` | Set `ForceMode` on all matches |
| `Sort()` | Sort alphabetically by first trigger |
| `Deduplicate()` | Remove matches with duplicate triggers |
| `Filter(func(Match) bool)` | Keep only matches where predicate is true |
| `Append(...Matches)` | Concatenate with other Matches |
| `Validate()` | Check all matches are well-formed |

All collection methods return a new `Matches` value -- the original is never modified.

## Development

```bash
make help     # Show all available targets
make check    # Run all quality gates (fmt, tidy, vet, lint, test)
make lint     # Run golangci-lint
make test     # Run tests with race detector
```

## License

[MIT License](LICENSE)
