# Building Gosper from Source

This guide covers building Gosper from source code, including prerequisites, platform-specific setup, and development workflow.

## Quick Build

```bash
# Clone the repository
git clone https://github.com/yourusername/gosper.git
cd gosper

# Build whisper.cpp dependency
make deps

# Build CLI
make build

# Run tests
make test
```

## Prerequisites

### Required

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **GCC/Clang** - C compiler for CGO
- **Make** - Build automation
- **Git** - Version control

### Platform-Specific Requirements

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y build-essential git

# Fedora/RHEL
sudo dnf install gcc gcc-c++ make git
```

#### macOS
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Or install via Homebrew
brew install gcc make
```

#### Windows
- Install [MSYS2](https://www.msys2.org/) or [MinGW-w64](https://www.mingw-w64.org/)
- Add GCC to PATH
- Use Git Bash or PowerShell

## Build Steps

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/gosper.git
cd gosper
```

### 2. Build whisper.cpp Dependency

Gosper uses [whisper.cpp](https://github.com/ggerganov/whisper.cpp) for speech recognition. Build the static library:

```bash
make deps
```

This will:
1. Clone whisper.cpp as a Git submodule
2. Build the C++ library
3. Build Go bindings

**Output**:
- Static library: `whisper.cpp/libwhisper.a`
- Go bindings: `whisper.cpp/bindings/go/build/`

### 3. Build Gosper

#### CLI Binary

Build the command-line interface:

```bash
make build
```

This creates `dist/gosper` with full features (CLI + Whisper + microphone support).

**Custom builds with build tags**:

```bash
# CLI + Whisper only (no microphone)
go build -tags "cli whisper" -o dist/gosper ./cmd/gosper

# CLI + Microphone only (no transcription)
go build -tags "cli malgo" -o dist/gosper ./cmd/gosper

# Minimal CLI (list devices only)
go build -tags "cli" -o dist/gosper ./cmd/gosper
```

#### Server Binary

Build the HTTP API server:

```bash
go build -tags "whisper" -o dist/server ./cmd/server
```

This creates a binary that exposes `/api/transcribe` endpoint.

### 4. Verify Build

Test the CLI:

```bash
# Show help
./dist/gosper --help

# Transcribe sample audio
./dist/gosper transcribe whisper.cpp/samples/jfk.wav \
  --model ggml-tiny.en.bin \
  --lang en
```

## Build Tags Explained

Gosper uses Go build tags to conditionally compile features. This reduces binary size and avoids unnecessary dependencies.

| Tag | Purpose | Dependencies | When to Use |
|-----|---------|-------------|-------------|
| `cli` | Enable CLI commands | None | Building command-line tool |
| `whisper` | Enable Whisper transcription | CGO, whisper.cpp | Need speech recognition |
| `malgo` | Enable microphone capture | CGO, miniaudio | Need audio recording |

### Common Build Combinations

```bash
# Full-featured CLI (recommended for development)
go build -tags "cli malgo whisper" -o gosper ./cmd/gosper

# HTTP Server (production backend)
go build -tags "whisper" -o server ./cmd/server

# Lightweight CLI (testing/debugging)
go build -tags "cli" -o gosper ./cmd/gosper
```

## Development Workflow

### Project Structure

```
gosper/
├── cmd/
│   ├── gosper/          # CLI entrypoint
│   └── server/          # HTTP server entrypoint
├── internal/
│   ├── domain/          # Business entities
│   ├── port/            # Interfaces
│   ├── usecase/         # Application logic
│   └── adapter/         # Implementations
├── web/                 # Frontend static files
├── whisper.cpp/         # Git submodule (C++ library)
├── Makefile             # Build automation
└── go.mod               # Go dependencies
```

### Makefile Targets

```bash
make deps        # Build whisper.cpp
make build       # Build CLI
make server      # Build HTTP server
make test        # Run unit tests
make itest       # Run integration tests
make coverage    # Generate coverage report
make clean       # Remove build artifacts
make help        # Show all targets
```

### Running Tests

**Unit Tests**:
```bash
make test
```

**Integration Tests** (requires Whisper model):
```bash
export GOSPER_INTEGRATION=1
export GOSPER_MODEL_PATH=/path/to/ggml-tiny.en.bin
make itest
```

**Coverage Report**:
```bash
make coverage
# Opens coverage.html in browser
```

### Code Generation

If you modify protobuf definitions or generate code:

```bash
# Install code generation tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate code
go generate ./...
```

## Docker Build

### Build Images

```bash
# Backend server
docker build -f Dockerfile.server -t gosper/server:local .

# Frontend
docker build -f Dockerfile.frontend -t gosper/fe:local .
```

### Multi-Stage Build Process

**Dockerfile.server**:
1. **Stage 1** (`builder`): Build whisper.cpp and Go binary
2. **Stage 2** (`runtime`): Minimal runtime image with binary only

**Benefits**:
- Small image size (~100MB vs 1GB+ with build tools)
- No build tools in production image
- Reproducible builds

### Development with Docker

```bash
# Run server locally
docker run -p 8080:8080 \
  -v $(pwd)/models:/models \
  -e GOSPER_MODEL=ggml-tiny.en.bin \
  gosper/server:local

# Test transcription
curl -F audio=@test.wav http://localhost:8080/api/transcribe
```

## Platform-Specific Build Notes

### Linux

**Audio Device Access**:
```bash
# Add user to audio group
sudo usermod -a -G audio $USER

# Reboot or re-login for changes to take effect
```

**Libraries**:
```bash
# Install audio libraries (for malgo support)
sudo apt-get install -y libasound2-dev  # ALSA
sudo apt-get install -y libpulse-dev    # PulseAudio
```

### macOS

**Microphone Permissions**:
- First run of `gosper record` triggers permission prompt
- Grant access in System Settings → Privacy & Security → Microphone
- If prompt doesn't appear, manually add Terminal/iTerm to allowed apps

**ARM64 (Apple Silicon)**:
```bash
# Build for native ARM64
GOARCH=arm64 make build

# Build for x86_64 (Rosetta)
GOARCH=amd64 make build

# Universal binary
lipo -create dist/gosper-arm64 dist/gosper-amd64 -output dist/gosper
```

### Windows

**CGO Setup**:
```bash
# Using MSYS2
pacman -S mingw-w64-x86_64-gcc

# Set environment
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
```

**Audio Device Access**:
- Windows Defender may prompt for microphone access
- Grant permission in Settings → Privacy → Microphone

**Build from PowerShell**:
```powershell
# Build deps
mingw32-make deps

# Build Gosper
$env:CGO_ENABLED=1
go build -tags "cli malgo whisper" -o gosper.exe ./cmd/gosper
```

## Troubleshooting

### CGO Errors

**Error**: `gcc: command not found`

**Solution**: Install GCC/Clang (see platform-specific requirements above)

---

**Error**: `undefined reference to whisper_*`

**Solution**: Build whisper.cpp first:
```bash
make deps
```

---

**Error**: `cannot find whisper.h`

**Solution**: Set include path:
```bash
export C_INCLUDE_PATH="$(pwd)/whisper.cpp:$(pwd)/whisper.cpp/bindings/go/build"
export LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"
```

### Build Tag Errors

**Error**: `undefined: cmd` when running binary

**Cause**: Missing `cli` build tag

**Solution**: Build with correct tags:
```bash
go build -tags "cli whisper" -o gosper ./cmd/gosper
```

### Linker Errors

**Error**: `ld: library not found for -lwhisper`

**Solution** (macOS):
```bash
export LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"
go build -tags "cli whisper" -o gosper ./cmd/gosper
```

**Solution** (Linux):
```bash
export LD_LIBRARY_PATH="$(pwd)/whisper.cpp/bindings/go/build"
go build -tags "cli whisper" -o gosper ./cmd/gosper
```

### Integration Test Failures

**Error**: `model not found`

**Solution**: Download model first or set path:
```bash
# Download tiny model
curl -L -o ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Set path for tests
export GOSPER_MODEL_PATH=$(pwd)/ggml-tiny.en.bin
make itest
```

## Advanced Topics

### Cross-Compilation

```bash
# Build for Linux from macOS
GOOS=linux GOARCH=amd64 go build -tags "whisper" -o server-linux ./cmd/server

# Build for Windows from Linux
GOOS=windows GOARCH=amd64 go build -tags "cli whisper" -o gosper.exe ./cmd/gosper
```

**Note**: Cross-compiling with CGO requires target platform's GCC toolchain.

### Static Linking

For fully static binaries (useful for Alpine Linux):

```bash
CGO_ENABLED=1 go build \
  -tags "whisper cli malgo" \
  -ldflags "-linkmode external -extldflags -static" \
  -o gosper ./cmd/gosper
```

### Debug Builds

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -tags "cli whisper" -o gosper ./cmd/gosper

# Run with delve debugger
dlv exec ./gosper -- transcribe test.wav
```

### Profiling

```bash
# Build with profiling
go build -tags "cli whisper" -o gosper ./cmd/gosper

# CPU profile
./gosper transcribe large-file.mp3 --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build deps
        run: make deps

      - name: Build
        run: make build

      - name: Test
        run: make test

      - name: Coverage
        run: make coverage
```

## Next Steps

- **[Quick Start Guide](QUICKSTART.md)** - Try Gosper with examples
- **[Architecture](ARCHITECTURE.md)** - Understand code structure
- **[Contributing](CONTRIBUTING.md)** - Development guidelines
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues
