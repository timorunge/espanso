// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

// Ci runs several mage tasks at once.
func Ci() error {
	if err := Fmt(); err != nil {
		return err
	}
	if err := Lint(); err != nil {
		return err
	}
	if err := StaticCheck(); err != nil {
		return err
	}
	if err := Test(); err != nil {
		return err
	}
	return nil
}

// Fmt runs gofmt over all go files.
func Fmt() error {
	files, err := goFiles()
	if err != nil {
		return err
	}
	failed := false
	for _, file := range files {
		_, err := sh.Output("gofmt", "-l", file)
		if err != nil {
			fmt.Printf("gofmt error on %s: %v", file, err)
			failed = true
		}
	}
	if failed {
		return errors.New("gofmt failed")
	}
	return nil
}

// Lint runns golangci-lint
func Lint() error {
	mg.Deps(getGoDependencies)
	return sh.Run("golangci-lint", "run", "-E", "misspell")
}

// StaticCheck runs staticcheck ./...
func StaticCheck() error {
	mg.Deps(getGoDependencies)
	return sh.Run("staticcheck", "./...")
}

// Test lazygithub.
func Test() error {
	return sh.Run("go", "test", "-v", "./...")
}

// getGoDependencies is installing all go dependency packages.
func getGoDependencies() error {
	if err := sh.Run("go", "get", "github.com/golangci/golangci-lint/cmd/golangci-lint"); err != nil {
		return errors.New("unable to install golangci-lint")
	}
	if err := sh.Run("go", "get", "honnef.co/go/tools/cmd/staticcheck"); err != nil {
		return errors.New("unable to install staticcheck")
	}
	return nil
}

// goFiles finds all go files.
func goFiles() ([]string, error) {
	goFiles := []string{}
	err := filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if ".go" == filepath.Ext(path) {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	return goFiles, err
}
