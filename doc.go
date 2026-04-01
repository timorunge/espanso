// Package espanso provides types and functions for generating espanso v2
// text-expander package files programmatically.
//
// Compatible with espanso v2 (tested against v2.2+). Covers the full match
// specification: all trigger types, output types, behavior modifiers,
// context filters, variables, and form fields.
//
// # Core types
//
// [Package] writes a complete espanso package (package.yml) with matches,
// global variables, and imports. Use [ReadPackageFile] and [ReadPackageDir]
// to read existing packages back.
//
// [MatchFile] writes standalone match files (match/*.yml) without the
// package wrapper -- useful for personal espanso configurations.
//
// [Manifest] writes hub registry metadata (_manifest.yml) required for
// publishing to the espanso hub.
//
// [Readme] and [License] generate the corresponding metadata files.
//
// # Convenience functions
//
// [WriteAll] writes package.yml, README.md, and LICENSE in one call.
// [WriteHubPackage] writes all files required for hub submission
// (_manifest.yml, package.yml, README.md) in the correct directory layout.
//
// # Quick start
//
//	p := espanso.Package{
//	    Name:    "my-package",
//	    Parent:  "default",
//	    Version: "0.1.0",
//	    Matches: espanso.Matches{
//	        {Triggers: []string{":hello"}, Replace: "Hello World"},
//	    },
//	}
//	if err := p.WriteFile("my-package/0.1.0"); err != nil {
//	    log.Fatal(err)
//	}
package espanso
