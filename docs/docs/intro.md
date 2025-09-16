# Welcome to Protodex

**Protodex** is a self-hosted, lightweight protobuf schema registry and code generation tool. It provides operations for managing, versioning, and distributing Protocol Buffer schemas.

## What is Protodex?

Protodex is designed to solve the challenges of managing protobuf schemas in distributed systems by providing:

- **Schema Registry** - Self-hosted registry for storing and versioning .proto files
- **Version Management** - Registry-like push/pull operations for schema distribution
- **Code Generation** - Generate client code for Go, Python, JavaScript/TypeScript, Java and more
- **Validation** - Built-in schema validation using protoc
- **Project Management** - YAML-based configuration for organizing proto files
- **Plugin System** - Automatic plugin management and custom protoc plugin support
- **Web Interface** - Browser-based interface for exploring packages and schemas

## Key Features

### **Self-Hosted**
- Run your own private registry server
- No external dependencies or cloud services required
- Simple binary deployment with built-in web interface

### **Multi-Language Support** 

- Generate code for Go, Python, JavaScript/TypeScript, Java
- Support for any protoc-compatible plugin (gRPC, Twirp, gRPC-Gateway, Validate, etc.)
- Customizable output directories

### **Package Management**
- Semantic versioning for schema packages
- Project bundling with directory structure preservation

### **Developer Experience**
- Simple `protodex.yaml` configuration
- Intuitive CLI commands (init, push, pull, generate)
- Built-in schema validation

## Quick Example

```bash


# Initialize a new protodex project
protodex init my-service

# Add your proto files
mkdir -p proto/user
cat > proto/user/user.proto << EOF
syntax = "proto3";
package user.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string user_id = 1;
  string name = 2;
  string email = 3;
}
EOF

# Generate code for Go
protodex generate go

# Start a local registry server
protodex serve --port 8080

# Make sure to sign up via the web interface at http://localhost:8080

# Push your package to the registry
protodex push v1.0.0

# Pull a package from the registry
protodex pull user-service:v1.0.0 ./schemas
```

## Get Started

Ready to start using Protodex? Here's what to do next:

1. **[Install Protodex](installation.md)** - Download and install the CLI tool
2. **[Quick Start](quick-start.md)** - Create your first project
3. **[Configuration](yaml-config.md)** - Learn about protodex.yaml configuration
4. **[CLI Reference](cli-overview.md)** - Master the command-line interface

## Community & Support

- **Source Code**: [github.com/sirrobot01/protodex](https://github.com/sirrobot01/protodex)
- **Issues & Bug Reports**: [GitHub Issues](https://github.com/sirrobot01/protodex/issues)

---

**Ready to get started?** Jump to the [Installation Guide](installation.md) and create your first protodex project!