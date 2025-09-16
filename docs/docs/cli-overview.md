# CLI Reference

Complete reference for the `protodex` command-line interface.

## Installation

The `protodex` CLI is a single binary that provides all functionality. See
the [Installation Guide](installation.md) for setup instructions.

## Global Options

Global flags available for all commands:

```bash
protodex [global options] command [command options] [arguments]
```

## Commands Overview

### Project Commands

| Command    | Description                         |
|------------|-------------------------------------|
| `init`     | Initialize a new protodex project   |
| `validate` | Validate protobuf schemas           |
| `generate` | Generate code from protobuf schemas |
| `deps`     | Manage project dependencies         |
| `source`   | Validate a source URL               |

### Registry Commands

| Command  | Description                  |
|----------|------------------------------|
| `login`  | Authenticate with a registry |
| `logout` | Clear authentication token   |
| `push`   | Push package to registry     |
| `pull`   | Pull package from registry   |

### Server Commands

| Command | Description                    |
|---------|--------------------------------|
| `serve` | Start protodex registry server |

### Configuration Commands

| Command  | Description               |
|----------|---------------------------|
| `config` | Show configuration values |

## Command Details

### `protodex init`

Initialize a new protodex project by creating a `protodex.yaml` configuration file.

**Usage:**

```bash
protodex init [package-name] [flags]
```

**Examples:**

```bash
protodex init                           # Interactive initialization
protodex init user-service             # Initialize with package name
protodex init --description "User API" # With description
protodex init user-service ./dir      # Specify project directory
```

**Flags:**

- `--description, -d` - Package description

**What it does:**

- Creates `protodex.yaml` configuration file
- Sets up basic project structure
- Configures default code generation for Go
- Adds Google protobuf dependency

---

### `protodex validate`

Validate protobuf schemas in the current project.

**Usage:**

```bash
protodex validate [flags]
```

**Examples:**

```bash
protodex validate                      # Validate all files in local project
```

**What it does:**

- Parses and validates proto syntax
- Checks import dependencies
- Validates against protobuf rules
- Reports errors and warnings

---

### `protodex generate [language]`

Generate code from protobuf schemas.

**Usage:**

```bash
protodex generate [language] [flags]
```

**Examples:**

```bash
protodex generate go                   # Generate only Go code
protodex generate python --output ./py # Generate Python with custom output
```

**Flags:**

- `--output, -o` - Override output directory
- `--clean` - Clean output directory before generating

**What it does:**
- 

- Manages protoc plugins automatically
- Resolves dependencies from configuration
- Generates code according to configuration and flags
- Supports multiple languages and plugins

---

### `protodex serve`

Start the protodex registry server with web interface.

**Usage:**

```bash
protodex serve [flags]
```

**Examples:**

```bash
protodex serve                         # Start on default port 3000
protodex serve --port 8080            # Start on custom port
protodex serve --data-dir ./registry  # Use custom data directory
```

**Flags:**

- `--port, -p` - Server port (default: 3000)
- `--data-dir` - Data directory for storage (default: ./data)

**What it does:**

- Starts HTTP server with REST API
- Serves web interface for browsing packages
- Handles user authentication and package storage
- Provides registry functionality for push/pull operations

---

### `protodex login`

Authenticate with a protodex registry server.

**Usage:**

```bash
protodex login [flags]
```

**Examples:**

```bash
protodex login                # Interactive login
protodex login --username admin --password secret
```

**Flags:**

- `--username, -u` - Username for authentication
- `--password, -p` - Password for authentication

**What it does:**

- Prompts for credentials if not provided
- Authenticates with registry server
- Stores authentication token in configuration
- Enables push/pull operations

---

### `protodex logout`

Clear stored authentication token.

**Usage:**

```bash
protodex logout
```

**What it does:**

- Removes stored authentication token
- Disables authenticated registry operations

---

### `protodex push`

Push a package version to the registry.

**Usage:**

```bash
protodex push <version> [flags]
```

**Examples:**

```bash
protodex push v1.0.0          # Push version v1.0.0
protodex push v1.2.0 ./my-project  # Push from specific directory
```

**What it does:**

- Validates project configuration and proto files
- Creates zip archive with project files and structure
- Includes `protodex.yaml`, proto files, and README.md (if present)
- Uploads package to registry with version metadata
- Maintains file directory structure

---

### `protodex pull`

Pull a package from the registry.

**Usage:**

```bash
protodex pull <package:version> [output-path] [flags]
```

**Examples:**

```bash
protodex pull user-service:v1.0.0         # Pull to current directory
protodex pull user-service:latest ./deps  # Pull to specific directory
```

**What it does:**

- Downloads package zip from registry
- Extracts files maintaining directory structure
- Preserves original project structure

---

### `protodex deps`

Manage project dependencies.
**Usage:**

```bash
protodex deps [command] [flags]
```

**Subcommands:**

- `list` - List current dependencies
- `add <name> <source>` - Add a new dependency
- `resolve` - Resolve and download dependencies

**Examples:**

```bash
protodex deps list                     # List dependencies
protodex deps add common/types protodex://common-times@0.0.4 --resolve # Add dependency from registry
protodex deps resolve                  # Resolve and download all dependencies
```

---

### `protodex source`

Validate a source URL for dependencies.
**Usage:**

```bash
protodex source <url> [flags]
```

**Examples:**

```bash
protodex source protodex://common/types@v1.0.0  # Validate protodex URL
protodex source github://user/repo@main          # Validate GitHub URL
protodex source file:///path/to/schemas        # Validate local path URL
```

---

### `protodex config`

Show current configuration values.

**Usage:**

```bash
protodex config
```

**What it shows:**

- Registry URL
- Authentication status
- User configuration
- Global settings

## Configuration File

Protodex stores global configuration in `~/.protodex/config.yaml`:

```yaml
registry: http://localhost:3000
hashed_token: <authentication-token>
```

## Common Workflows

### Creating a New Project

```bash
# Initialize project
protodex init my-service --description "My service API"

# Add proto files
mkdir -p proto/user
cat > proto/user/user.proto << EOF
syntax = "proto3";
package user.v1;
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
message GetUserRequest { string user_id = 1; }
message GetUserResponse { string user_id = 1; string name = 2; }
EOF

# Generate code
protodex generate go

# Validate
protodex validate
```

### Publishing to Registry

```bash
# Start local registry (separate terminal)
protodex serve --port 8080

# Login to registry
protodex login

# Push package
protodex push v1.0.0
```

## Getting Help

Get help for any command:

```bash
protodex --help                    # General help
protodex init --help              # Command-specific help
protodex push --help     # Subcommand help
```