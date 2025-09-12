# Configuration Reference

Complete reference for the `protodex.yaml` configuration file.

## Basic Structure

The `protodex.yaml` file defines your protobuf project configuration:

```yaml
package:
  name: my-service
  description: My service protobuf schemas

files:
  exclude: []
  base_dir: "."

gen:
  languages:
    - name: go
      output_dir: ./gen/go

deps:
  - name: google/protobuf
    source: google

plugins:
  - name: grpc-gateway
    command: protoc-gen-grpc-gateway
    output_dir: grpc-gateway_out
    options:
      logtostderr: "true"
```

## Package Configuration

Define your package metadata:

```yaml
package:
  name: user-service              # Required: Package name for registry
  description: User service API   # Optional: Package description
```

### Version Format
- Must follow semantic versioning (semver)
- Can optionally start with 'v': `v1.0.0` or `1.0.0`
- Examples: `v1.0.0`, `v2.1.0-beta.1`, `1.2.3`


## File Configuration

Control which proto files are included in your project:

```yaml
files:
  exclude:                    # File patterns to exclude (optional)
    - "**/*_test.proto"      # Exclude test files
    - "proto/internal/"      # Exclude internal directory
  
  base_dir: "."              # Base directory for file resolution
```

## Code Generation

Configure code generation for multiple languages:

### Go

```yaml
gen:
  languages:
    - name: go
      output_dir: ./gen/go
      options:
        go_opt: paths=source_relative
        
      plugins:
        - name: grpc-gateway
          command: protoc-gen-grpc-gateway
          output_dir: grpc-gateway_out
          options:
            logtostderr: "true"
```

### Python

```yaml
gen:
  languages:
    - name: python
      output_dir: ./gen/python
      
      options:
        python_opt: ""
```

### JavaScript/TypeScript

```yaml
gen:
  languages:
    - name: js
      output_dir: ./gen/js
      
      options:
        js_out: "import_style=commonjs,binary"
        
    - name: ts
      output_dir: ./gen/ts
      
      options:
        ts_out: "generate_package_definition"
```

### Java

```yaml
gen:
  languages:
    - name: java
      output_dir: ./gen/java
      
      options:
        java_opt: ""
```

## Dependencies

Manage external protobuf dependencies:

```yaml
deps:
  - name: google/protobuf          # Google's well-known types
    type: google-well-known      # Use built-in source
    
  - name: common/types     # Custom dependency from registry
    type: protodex               # Use protodex registry
    version: v1.2.0               # Specific version
    source: common-types
    
  - name: external/schemas         # GitHub dependency  
    type: github                 # Use GitHub as source
    source: user/repo         # GitHub repo (user/repo)
    version: main                # Branch, tag, or commit

  - name: local/schemas           # Local filesystem dependency
    type: local                  # Use local path
    source: /path/to/schemas  # Local directory path
    version: ""                  # Optional version (for consistency)
```

### Dependency Sources

- `google` - Google's well-known types (automatically resolved)
- `protodex` - Pull from protodex registry
- `github` - Pull from GitHub repository
- `local` - Local filesystem path (for development)

## Plugin Configuration

Configure protoc plugins for code generation:

### Global Plugins

Applied to all languages:

```yaml
plugins:
  - name: validate                    # Plugin name
    command: protoc-gen-validate      # Executable name
    output_dir: validate_out              # Output flag for protoc
    required: true                    # Must be installed
    options:
      lang: go                       # Plugin-specific options
```

### Language-Specific Plugins

Applied only to specific languages:

```yaml
gen:
  languages:
    - name: go
      output_dir: ./gen/go
      plugins:
        - name: twirp
          command: protoc-gen-twirp
          output_dir: twirp_out
          grpc: false
          options:
            package_prefix: github.com/myorg/my-service
```

### Plugin Properties

- `name` - Human-readable plugin name
- `command` - Executable name (must be in PATH or managed by protodex)
- `output_dir` - Output flag passed to protoc (e.g., `./gen/twirp`)
- `required` - Whether plugin must be installed (default: false)
- `options` - Key-value pairs passed to the plugin

## Complete Example

```yaml
# Package metadata
package:
  name: user-service
  description: User management service API

# File configuration
files:
  exclude:
    - "proto/internal/**"
  base_dir: "."

# Code generation
gen:
  languages:
    # Go with gRPC
    - name: go
      output_dir: ./gen/go
      
      options:
        go_opt: paths=source_relative
      plugins:
        - name: grpc-gateway
          command: protoc-gen-grpc-gateway
          output_dir: ./grpc-gateway_out
          options:
            logtostderr: "true"
            
    # Python with gRPC
    - name: python
      output_dir: ./gen/python  
      

# Dependencies
deps:
  - name: google/protobuf          # Google's well-known types
    type: google-well-known      # Use built-in source

  - name: common/types     # Custom dependency from registry
    type: protodex               # Use protodex registry
    version: v1.2.0               # Specific version
    source: common-types

# Global plugins
plugins:
  - name: validate
    command: protoc-gen-validate
    output_dir: validate_out
    required: true
    options:
      lang: go
```

## Validation

Validate your project configuration and proto files:

```bash
protodex validate
```

This checks:
- YAML syntax
- Required fields
- File path existence
- Plugin availability
- Dependency resolution