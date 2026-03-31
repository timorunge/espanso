// Package espanso provides types and functions for generating espanso
// text-expander package files: package.yml, README.md, and LICENSE.
//
// Typical usage:
//
//	p := espanso.Package{
//	    Name:    "my-package",
//	    Parent:  "default",
//	    Version: "0.1.0",
//	    Matches: espanso.DictToMatches(dict),
//	}
//	if err := p.WriteFile(); err != nil { ... }
package espanso
