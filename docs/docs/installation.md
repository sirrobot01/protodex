# Installation

Get Protodex up and running on your system.

## Install Binary

### Using Go

Install the latest version directly from GitHub:

```bash
go install github.com/sirrobot01/protodex/cmd/protodex@latest
```

This installs the `protodex` binary to your `$GOPATH/bin` (typically `~/go/bin`).

### Pre-built Binaries

Download pre-built binaries from the [GitHub Releases](https://github.com/sirrobot01/protodex/releases) page:

```bash
# Linux
curl -LO https://github.com/sirrobot01/protodex/releases/latest/download/protodex-linux-amd64.tar.gz
tar -xzf protodex-linux-amd64.tar.gz
sudo mv protodex /usr/local/bin/

# macOS
curl -LO https://github.com/sirrobot01/protodex/releases/latest/download/protodex-darwin-amd64.tar.gz
tar -xzf protodex-darwin-amd64.tar.gz
sudo mv protodex /usr/local/bin/

# Windows
# Download protodex-windows-amd64.zip and extract to PATH
```

## Protoc Compiler

Protodex requires the Protocol Buffers compiler `protoc` to generate code from `.proto` files.

It automatically downloads a compatible version of `protoc` on first run. This is stored in `~/.protodex/bin`. 

### Using your own protoc

You can use your own version of `protoc` if desired by changing the your ~/.protodex/config.yaml file:

```yaml
protoc:
  bin: /path/to/your/protoc
```

## Configuration

Protodex stores global configuration in `~/.protodex/config.yaml`. The file is created automatically on first use.

View current configuration:
```bash
protodex config
```

## Troubleshooting

### Command Not Found

If you see `command not found: protodex`:

1. **Check Installation**: Verify the binary is installed
   ```bash
   which protodex
   ls ~/go/bin/protodex
   ```

2. **Check PATH**: Ensure `~/go/bin` is in your PATH
   ```bash
   echo $PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

3. **Reload Shell**: Restart your terminal or reload shell configuration
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

## Next Steps

Now that you have Protodex installed:

1. **[Quick Start](quick-start)** - Create your first project
2. **[Configuration](yaml-config.md)** - Learn about project configuration
3. **[CLI Reference](overview.md)** - Master the command-line interface

## Getting Help

If you encounter issues:

- **Documentation**: Check the [CLI Reference](overview.md)
- **GitHub Issues**: Report bugs at [github.com/sirrobot01/protodex/issues](https://github.com/sirrobot01/protodex/issues)