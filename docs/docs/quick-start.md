# Quick Start

Get up and running with Protodex in minutes. This guide will walk you through creating your first project, generating code, and using the registry.

## Step 1: Initialize a New Project

Create a new directory and initialize your protodex project:

```bash
mkdir my-protodex-project
cd my-protodex-project
protodex init my-service
```

This creates a `protodex.yaml` configuration file with default settings:

```yaml
package:
  name: my-service
  description: ""

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
```

## Step 2: Create Your First Proto File

Create a directory structure for your proto files:

```bash
mkdir -p proto/user
```

Create `proto/user/user.proto`:

```protobuf
syntax = "proto3";

package user.v1;

option go_package = "github.com/myorg/my-service/gen/go/user/v1";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
  string created_at = 4;
}

message CreateUserRequest {
  string name = 1;
  string email = 2;
}

message CreateUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
  string created_at = 4;
}
```

## Step 3: Configure Code Generation

Edit `protodex.yaml` to configure code generation for multiple languages:

```yaml
package:
  name: my-service
  description: My first protodex service

files:
  base_dir: "."

gen:
  languages:
    - name: go
      output_dir: ./gen/go
      
    
    - name: python
      output_dir: ./gen/python
      

deps:
  - name: google/protobuf
    source: google
```

## Step 4: Validate Your Schemas

Ensure your proto files are valid:

```bash
protodex validate
```

You should see output like:
```
Found 1 proto files
Validating proto files
Validation successful
```

## Step 5: Generate Code

Generate code for your configured languages:

```bash
protodex generate go
```

This will:
1. Download necessary protoc plugins automatically
2. Resolve dependencies (like Google's protobuf types)
3. Generate code in the specified output directories

Your project structure will now look like:

```
my-protodex-project/
├── protodex.yaml
├── proto/
│   └── user/
│       └── user.proto
└── gen/
    ├── go/
    │   └── user/
    │       └── v1/
    │           ├── user.pb.go
    │           └── user_grpc.pb.go
    └── python/
        └── user/
            └── user_pb2.py
            └── user_pb2_grpc.py
```

## Step 6: Set Up a Registry (Optional)

Start a local registry server to share your schemas:

```bash
# In a separate terminal
protodex serve --port 8080
```

The registry provides:
- Web interface at `http://localhost:8080`
- REST API at `http://localhost:8080/api`

## Step 7: Register and Push Your Package

Register a user account:

```bash
protodex login
# Follow prompts to create account or login
```

Push your package to the registry:

```bash
protodex push v0.1.0
```

This will:
1. Create a ZIP archive with your project files
2. Include `protodex.yaml`, proto files, and README.md (if present)
3. Upload to the registry maintaining directory structure

## Step 8: Pull and Use Dependencies

Try pulling your package from the registry:

```bash
# In another directory
mkdir test-pull && cd test-pull
protodex pull my-service:v0.1.0
```

## Plugins

Protodex uses `protoc` plugins to generate code for different languages and frameworks. Plugins required for simple message code generation are downloaded automatically as needed.

For other plugins (e.g. gRPC, validation), you may need to install them manually and ensure they are in your PATH. You can then configure them in your `protodex.yaml` file. Check the [CLI Reference](cli-overview.md) for details.

```yaml
gen:
  languages:
    - name: go
      output_dir: ./gen/go
      
      plugins:
        - name: grpc-gateway
          command: protoc-gen-grpc-gateway
          output_dir: grpc-gateway_out
          options:
            logtostderr: "true"
```

Install the plugin:
```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

## Common Workflows

### Development Workflow

```bash
# 1. Make changes to proto files
vim proto/user/user.proto

# 2. Validate changes
protodex validate

# 3. Regenerate code
protodex generate go

# 4. Test your application
go run main.go
```

### Release Workflow

```bash
# 1. Update version
vim protodex.yaml  # Update package.version

# 2. Validate and generate
protodex validate
protodex generate go

# 3. Push to registry
protodex push v1.0.0

# 4. Tag release
git tag v1.0.0
git push origin v1.0.0
```

## Troubleshooting

### Plugin Not Found

```bash
# Make sure the plugin is installed and in your PATH
which protoc-gen-grpc-gateway

which protoc-gen-{name_of_plugin}
```

## Next Steps

Now that you've created your first Protodex project:

1. **[Configuration Guide](yaml-config.md)** - Learn about advanced configuration options
2. **[CLI Reference](cli-overview.md)** - Master all available commands
3. **[Registry Guide](cli-overview.md)** - Set up and manage registries

## Getting Help

If you run into issues:

- **CLI Help**: Run `protodex --help` or `protodex <command> --help`
- **Validation**: Use `protodex validate` to check your setup
- **GitHub Issues**: Report problems at [github.com/sirrobot01/protodex/issues](https://github.com/sirrobot01/protodex/issues)