// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"time"
)

const (
	packageName = "github.com/g1ntas/accio/cmd/accio"
)

// Builds Accio binary
func Build() error {
	return sh.RunV(mg.GoCmd(), "build", "-ldflags", ldflags(), packageName)
}

// Installs Accio binary
func Install() error {
	return sh.RunV(mg.GoCmd(), "install", "-ldflags", ldflags(), packageName)
}

// Runs tests and linters
func Check() error {
	mg.Deps(Lint)
	mg.Deps(TestRace)
	return nil
}

// Runs tests
func Test() error {
	if os.Getenv("CI") != "" {
		return sh.RunV(mg.GoCmd(), "test", "-coverprofile=coverage.txt", "-covermode=atomic", "./...")
	}
	return sh.RunV(mg.GoCmd(), "test", "./...")
}

// Runs tests with race detector
func TestRace() error {
	if os.Getenv("CI") != "" {
		return sh.RunV(mg.GoCmd(), "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic", "./...")
	}
	return sh.RunV(mg.GoCmd(), "test", "-race", "./...")
}

// Runs golangci linter
func Lint() error {
	return sh.RunV("golangci-lint", "run", "./...")
}

// Reports test coverage
func Coverage() error {
	err := sh.Run(mg.GoCmd(), "test", "-v", "-coverprofile=coverage.out", "./...")
	if err != nil {
		return err
	}
	err = sh.RunV(mg.GoCmd(), "tool", "cover", "-func=coverage.out")
	if err != nil {
		return err
	}
	err = sh.Run(mg.GoCmd(), "tool", "cover", "-html=coverage.out")
	if err != nil {
		return err
	}
	return sh.Rm("coverage.out")
}

func ldflags() string {
	timestamp := time.Now().Format(time.RFC3339)
	hash := commit()
	tag := tag()
	if tag == "" {
		tag = "dev"
	}
	return fmt.Sprintf(`-X 'main.buildDate=%s' -X 'main.buildCommit=%s' -X 'main.buildTag=%s'`, timestamp, hash, tag)
}

// tag returns the git tag for the current branch or "" if none.
func tag() string {
	s, _ := sh.Output("git", "describe", "--tags")
	print(s)
	return s
}

// commit returns the git hash for the current repo or "" if none.
func commit() string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return hash
}
