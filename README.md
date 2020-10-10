![Accio](docs/assets/demo.gif)

----

[![Linux, macOS and Windows Build Status](https://travis-ci.org/g1ntas/accio.svg?branch=master)](https://travis-ci.org/g1ntas/accio)
[![Go Report Card](https://goreportcard.com/badge/github.com/g1ntas/accio)](https://goreportcard.com/report/github.com/g1ntas/accio)
[![codecov](https://codecov.io/gh/g1ntas/accio/branch/master/graph/badge.svg)](https://codecov.io/gh/g1ntas/accio)





## About
Accio is a scaffolding tool for generating boilerplate code, written in Golang. You can use it to create templates for repetitive code patterns and generate them interactively whenever you need them. The best part about it is that Accio allows you to customize most aspects of code generation with custom scripts.

## Features
* **Interactive data prompts** - configure prompts and use data in templates;   
* **Scripting** - write custom scripts to process input data;
* **Remote generators** - execute generators directly from Git repositories;
* **No external dependencies** - no need to install external applications, dependency managers, or other tools - everything works out of the box with a single binary;
* **Cross-platform** - builds for Linux, OS X, Windows, and others.

## Installation
### Pre-compiled binaries
**Homebrew**:  
`brew install g1ntas/tap/accio`

**Shell script**:  
`curl -sfL https://raw.githubusercontent.com/g1ntas/accio/master/install.sh | sh`

**Manually**:  
Download the pre-compiled binaries from [releases page](https://github.com/g1ntas/accio/releases).

### Building from source
To build a binary from the source code, you need to have [Go](https://golang.org/) installed first.

**Steps**:
1. Clone repository: `git clone https://github.com/g1ntas/accio`
2. Build: `go run mage.go build`
3. Run `./accio` to verify if it works

## Quickstart
### Running generator
You can run a generator from a local directory with `run` command:

`accio run ./generator-directory`

Or directly from Git repository:

`accio run github.com/user/accio-generator-repo`

Subdirectories are supported as well:

`accio run github.com/g1ntas/accio/examples/open-source-license`

### Creating first generator
1. Create an empty config file `~/example/.accio.toml`

2. Create a template file `~/example/file.txt` with any content:
```
Hello, world
```

And that's all it takes to create a simple generator - now you can run it:
```
> accio run ~/example
$ Running...
$ Done.
> cat ~/example/TEST.TXT
$ Hello, world
```

To learn about more advanced features needed to write more complex generators, 
read [the introduction tutorial](docs/introduction.md).   

## Examples
* [github.com/g1ntas/accio/examples/go-travisci-config](examples/go-travisci-config) - generates TravisCI config with selected Go versions, operating systems, and CPU architectures
* [github.com/g1ntas/accio/examples/golang-cli-project](examples/golang-cli-project) - generates a boilerplate Go project with a Cobra CLI command
* [github.com/g1ntas/accio/examples/open-source-license](examples/open-source-license) - generates the selected open-source license

## Documentation
* [Introduction](docs/introduction.md)
* Core concepts
	* [Generators](docs/concepts/generators.md)
	* [Templates](docs/concepts/templates.md)
* Reference
	* [Blueprints](docs/reference/blueprints.md)
	* [Configuration](docs/reference/configuration.md)
	* [Markup language](docs/reference/accio-ml.md)
	* [Starlark](docs/reference/starlark.md)

## Contributing
Contributions are more than welcome, if you are interested please take a look to our [Contributing Guidelines](CONTRIBUTING.md).

## Copyright
Accio is released under the MIT license. See [LICENSE](LICENSE).