# Registry Server

The Protodex registry is a self-hosted solution for storing, versioning, and distributing protobuf schemas. It provides both a REST API and web interface for package management.

## What is the Registry?

The Protodex registry provides:

- **Package Storage** - Store and version protobuf schemas as packages
- **Dependency Management** - Resolve and download package dependencies
- **Web Interface** - Browser-based package exploration and management
- **Authentication** - User registration and login system
- **File Structure Preservation** - Maintains original project directory structure

## Starting the Registry

Start a local registry server:

```bash
protodex serve                    # Default port 3000
protodex serve --port 8080       # Custom port
protodex serve --data-dir ./data # Custom data directory
```

The server provides:
- REST API at `http://localhost:3000/api`
- Web interface at `http://localhost:3000`

## Package Structure

Registry packages are ZIP archives containing:

```
my-service-v1.0.0.zip
├── protodex.yaml        # Project configuration (required)
├── proto/              # Protocol buffer files
│   ├── user/
│   │   └── user.proto
│   └── auth/
│       └── auth.proto
├── README.md           # Documentation (optional)
└── other files...      # Additional project files
```

## Package Operations

### Push Package

Upload a package version to the registry:

```bash
protodex push v1.0.0
```

**What happens:**
1. Validates project configuration and proto files
2. Creates ZIP archive with all project files
3. Includes `protodex.yaml`, proto files, and README.md
4. Maintains directory structure in the archive
5. Uploads to registry with version metadata

### Pull Package

Download a package from the registry:

```bash
protodex pull user-service:v1.0.0
protodex pull user-service:latest ./deps
```

**What happens:**
1. Downloads ZIP archive from registry
2. Extracts files maintaining original directory structure
3. Preserves project layout and organization

## Versioning

Packages use semantic versioning:

- **Format**: `v1.0.0` or `1.0.0`
- **Examples**: `v1.2.3`, `v2.0.0-beta.1`, `1.0.0`
- **Special**: `latest` refers to most recent version

## Web Interface

The web interface provides:

### Package Browser
- Browse all available packages
- Search packages by name and description
- View package details and versions
- Access package documentation

### Package Details
- View schema files with syntax highlighting
- Browse file structure and contents
- See version history and metadata
- Download package versions


### Package Management
- Create new packages through web UI
- Upload package versions
- Manage package visibility

## Authentication

### User Registration

Register through web interface:

```bash
# Via web interface
# Navigate to http://localhost:3000 and click Register
```

### Login

Authenticate to enable push operations:

```bash
protodex login --username your-username
# Enter password when prompted
```

Authentication tokens are stored in `~/.protodex/config.yaml`.

### Logout

Clear stored authentication:

```bash
protodex logout
```

## Storage

### File Storage

Packages are stored in the data directory:

```
data/
├── schemas/
│   └── user-service/
│       └── v1.0.0/
│           ├── user-service-v1.0.0.zip
│           ├── proto/
│           │   └── user.proto
│           └── protodex.yaml
└── protodex.db
```

### Database

SQLite database stores:
- User accounts and authentication
- Package metadata
- Version information
- File references

# Command line flags
protodex serve --port 8080 --data-dir ./registry-data
```