# Protodex

[![Test](https://github.com/sirrobot01/protodex/actions/workflows/test.yml/badge.svg)](https://github.com/sirrobot01/protodex/actions/workflows/test.yml)
[![Release](https://github.com/sirrobot01/protodex/actions/workflows/release.yml/badge.svg)](https://github.com/sirrobot01/protodex/actions/workflows/release.yml)
[![Deploy Documentation](https://github.com/sirrobot01/protodex/actions/workflows/docs.yml/badge.svg)](https://github.com/sirrobot01/protodex/actions/workflows/docs.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sirrobot01/protodex.svg)](https://pkg.go.dev/github.com/sirrobot01/protodex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Release](https://img.shields.io/github/v/release/sirrobot01/protodex)](https://github.com/sirrobot01/protodex/releases)


## Overview

Protodex is a CLI tool and registry server for managing protobuf schemas and generating client code. It provides a simple workflow to push/pull protobuf packages to/from a central registry, validate schemas, and generate code in multiple languages using custom plugins.

You can view the documentation at [protodex.dev](https://protodex.dev).

## Features

- **Full-Featured CLI** - Familiar push/pull workflow for schema management
- **Built-in validation** - Protobuf syntax validation using protoc
- **Code generation** - Generate Go, Python, TypeScript, Java etc. clients
- **Custom plugins** - Support for any protoc plugin (GRPC, Twirp, gRPC-Gateway, Validate, etc.)

## Quick Start

### Installation

#### Download pre-built binaries:

Visit the [releases page](https://github.com/sirrobot01/protodex/releases) for your OS.


#### Build from source:
```bash
git clone https://github.com/sirrobot01/protodex
cd protodex
go build -o bin/protodex ./cmd/protodex
```

#### Install via Go:
```bash
go install github.com/sirrobot01/protodex/cmd/protodex@latest
```


### Registry Setup

The registry server is where you push and pull your protobuf packages. It's optional if you only need local code generation.

Start a local registry server (default port 8080):

```bash
protodex serve --port 8080 --data-dir ./data # data-dir defaults to ~/.protodex/data
```
The server provides:

- Web interface at `http://localhost:8080`
- REST API at `http://localhost:8080/api`

It **currently** uses a local SQLite database for now. I plan to add support for external databases in the future.

## Configuration

Create a `protodex.yaml` in your project root:

```bash
protodex init my-service
```

### Project Configuration (protodex.yaml)
```yaml
package:
  name: user-service
  description: User management service schemas

files:
  exclude: []
  base_dir: .

gen:
  languages:
    - name: go
      output_dir: ./gen/go
      module_path: github.com/myuser/user-service

deps:
  - name: google/protobuf
    type: google-well-known
    source: ""
```

### Custom Plugin Configuration

Protodex supports any protoc plugin through project configuration:

```yaml
# protodex.yaml

# Global plugins (applied to all languages)
plugins:
  - name: validate
    command: protoc-gen-validate
    required: false
    options:
      lang: go

gen:
  languages:
    - name: go
      output_dir: ./gen/go
      module_path: github.com/myuser/client
      # Language-specific plugins
      plugins:
        - name: twirp
          command: protoc-gen-twirp
          required: false
          options:
            package_prefix: github.com/myuser/client
    
    - name: python
      output_dir: ./gen/python
      
    - name: java
      output_dir: ./gen/java
      plugins:
        - name: grpc-gateway
          command: protoc-gen-grpc-gateway
          required: false
```

**Plugin Configuration Fields:**
- `name`: Plugin identifier
- `command`: Plugin executable name (e.g., `protoc-gen-twirp`)
- `output_dir`: Optional output directory for plugin-generated files
- `required`: Whether plugin must be installed (validates before generation)
- `options`: Plugin-specific options

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request


## TODO

- Support external databases (PostgreSQL, MySQL)
- Support for a better file storage backend (S3, GCS)
- Add a more robust authentication system
- Improve web interface with package browsing and search
- A more advanced plugin management system
- Support for custom generators

## License

MIT License - see [LICENCE](LICENCE) file for details.