// Convenience functions for writing all espanso package files at once.

package espanso

import (
	"errors"
	"fmt"
	"path/filepath"
)

// WriteAll writes package.yml, README.md, and LICENSE under rootDir.
// package.yml goes to rootDir/pkg.Name/pkg.Version/package.yml.
// README.md and LICENSE go to rootDir/pkg.Name/.
func WriteAll(rootDir string, pkg Package, r Readme, l License) error {
	var errs []error
	if err := pkg.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("package: %w", err))
	}
	if err := r.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("readme: %w", err))
	}
	if err := l.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("license: %w", err))
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	metaDir := filepath.Join(rootDir, pkg.Name)
	pkgDir := filepath.Join(metaDir, pkg.Version)

	if err := pkg.WriteFile(pkgDir); err != nil {
		return err
	}
	if err := r.WriteFile(metaDir); err != nil {
		return err
	}
	return l.WriteFile(metaDir)
}

// WriteHubPackage writes all files required for the espanso hub under rootDir.
// Creates rootDir/manifest.Name/manifest.Version/ with _manifest.yml,
// package.yml, and README.md.
func WriteHubPackage(rootDir string, manifest Manifest, pkg Package, r Readme) error {
	var errs []error
	if err := manifest.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("manifest: %w", err))
	}
	if err := pkg.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("package: %w", err))
	}
	if err := r.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("readme: %w", err))
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	pkgDir := filepath.Join(rootDir, manifest.Name, manifest.Version)

	if err := manifest.WriteFile(pkgDir); err != nil {
		return err
	}
	if err := pkg.WriteFile(pkgDir); err != nil {
		return err
	}
	return r.WriteFile(pkgDir)
}
