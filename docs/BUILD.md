# Building Gosper from Source

This guide covers building Gosper from source code, including prerequisites, platform-specific setup, and development workflow.

## Quick Build

```bash
# Clone the repository
git clone https://github.com/yourusername/gosper.git
cd gosper

# Build all binaries
make build-all

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

### 2. Build Dependencies and Binaries

```bash
make build-all
```

This will:
1. Initialize the `whisper.cpp` git submodule
2. Tidy the Go modules
3. Build the `whisper.cpp` static library
4. Build the CLI and server binaries

### 3. Verify Build

First, download a model to use for transcription:
```bash
# Build the model downloader utility
make -C whisper.cpp/bindings/go examples

# Download the tiny English model
./whisper.cpp/bindings/go/build_go/go-model-download -out whisper.cpp/models ggml-tiny.en.bin
```

Then, test the CLI:
```bash
# Show help
./dist/gosper --help

# Transcribe sample audio (use a WAV file)
./dist/gosper transcribe whisper.cpp/samples/jfk.wav \
  --model whisper.cpp/models/ggml-tiny.en.bin
```

## Build Tags Explained

Gosper uses Go build tags to conditionally compile features. This reduces binary size and avoids unnecessary dependencies.

| Tag | Purpose | Dependencies | When to Use |
|-----|---------|-------------|-------------|
| `cli` | Enable CLI commands | None | Building command-line tool |
| `whisper` | Enable Whisper transcription | CGO, whisper.cpp | Need speech recognition |
| `malgo` | Enable microphone capture | CGO, miniaudio | Need audio recording |

## Development Workflow

### Makefile Targets

```bash
make deps        # Build whisper.cpp
make tidy        # Tidy go modules
make build-all   # Build all binaries
make build-cli   # Build CLI binary
make build-server# Build server binary
make test        # Run unit tests
make itest       # Run integration tests
make lint        # Run linter
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
GOARCH=arm64 make build-all

# Build for x86_64 (Rosetta)
GOARCH=amd64 make build-all
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

**Solution**: This error typically occurs when building without using the main `Makefile`. The `Makefile` is configured to pass the correct include paths to the Go compiler.

**Recommended Fix**: Always use `make build-all` or other `make` targets to build the project.

If you must build manually with `go build`, you are responsible for setting the `C_INCLUDE_PATH` environment variable correctly. The `Makefile` handles this for you.

### Linker Errors

**Error**: `ld: library not found for -lwhisper`

**Solution**: The `Makefile` should handle this automatically. If you are building manually, you will need to set the appropriate library path environment variable for your operating system (e.g., `LIBRARY_PATH` on macOS, `LD_LIBRARY_PATH` on Linux).

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
GOOS=linux GOARCH=amd64 make build-server

# Build for Windows from Linux
GOOS=windows GOARCH=amd64 make build-cli
```

**Note**: Cross-compiling with CGO requires target platform's GCC toolchain.

### Static Linking

For fully static binaries (useful for Alpine Linux):

```bash
CGO_ENABLED=1 make build-all LDFLAGS="-linkmode external -extldflags -static"
```

### Debug Builds

```bash
# Build with debug symbols using Makefile (recommended)
make build-cli GOFLAGS='-gcflags="all=-N -l"'

# Or build manually with CGO vars
CGO_CFLAGS="-I$(pwd)/whisper.cpp/include:$(pwd)/whisper.cpp/ggml/include" \
CGO_LDFLAGS="-L$(pwd)/whisper.cpp/build_go/src -lwhisper -lggml -lggml-base -lggml-cpu -lm -lstdc++ -fopenmp" \
  go build -gcflags="all=-N -l" -tags "cli malgo whisper" -o dist/gosper ./cmd/gosper

# Run with delve debugger
dlv exec ./dist/gosper -- transcribe test.wav
```

### Profiling

```bash
# Build with profiling
make build-cli

# CPU profile
./dist/gosper transcribe large-file.mp3 --cpuprofile=cpu.prof

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

      - name: Build and test
        run: make build-all test
```

## Next Steps

- **[Quick Start Guide](QUICKSTART.md)** - Try Gosper with examples
- **[Architecture](ARCHITECTURE.md)** - Understand code structure
- **[Contributing](CONTRIBUTING.md)** - Development guidelines
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues
