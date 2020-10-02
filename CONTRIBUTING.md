# Contributing to Accio

When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change.

## Code of Conduct

This project and its contributors are expected to uphold the [Go Community Code of Conduct](https://golang.org/conduct). By participating, you are expected to follow these guidelines.

## Setup

Prerequisites:
* [Go 1.14+](https://golang.org/doc/install)
* [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)

Clone Accio:
```
git clone git@github.com:g1ntas/accio.git
```

Run tests and linters to verify everything is working:
```
go run mage.go check
```

Build:
```
go run mage.go build
```

## How to Contribute

In order for a PR to be accepted it needs to pass a list of requirements:

- All PRs must be written in idiomatic Go, formatted according to [gofmt](https://golang.org/cmd/gofmt/), and without any warnings from [go lint](https://github.com/golang/lint), [go vet](https://golang.org/cmd/vet), nor [golangci-lint](https://golangci-lint.run).
- They should in general include tests, and those shall pass.
- If the PR is a bug fix, it has to include a suite of unit tests and documentation for the new functionality.
- If the PR is a new feature, it has to come with a suite of unit tests and documentation for the new functionality.
- In any case, all the PRs have to pass the personal evaluation of at least one of the maintainers of Accio.
