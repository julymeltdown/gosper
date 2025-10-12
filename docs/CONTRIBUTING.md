# Contributing to Gosper

Thank you for your interest in contributing to Gosper! This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Submitting Changes](#submitting-changes)
- [Architecture Guidelines](#architecture-guidelines)
- [Documentation](#documentation)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inspiring community for all.

**Expected Behavior**:
- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on what is best for the community
- Show empathy towards other community members

**Unacceptable Behavior**:
- Harassment, discrimination, or derogatory comments
- Trolling, insulting comments, or personal attacks
- Publishing others' private information
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Instances of unacceptable behavior may be reported to the project maintainers. All complaints will be reviewed and investigated promptly and fairly.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- GCC or Clang (for CGO)
- Git
- Make

See [BUILD.md](BUILD.md) for detailed setup instructions.

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/gosper.git
cd gosper

# Add upstream remote
git remote add upstream https://github.com/ORIGINAL_OWNER/gosper.git
```

### Build Dependencies

```bash
# Build whisper.cpp
make deps

# Build Gosper
make build

# Run tests
make test
```

### Development Setup

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set environment for development
export GOSPER_LOG=debug
export GOSPER_MODEL=ggml-tiny.en.bin
```

## Development Workflow

### 1. Create a Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-123
```

### 2. Make Changes

- Write code following our [code standards](#code-standards)
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes

```bash
# Run unit tests
make test

# Run integration tests (requires model)
export GOSPER_INTEGRATION=1
export GOSPER_MODEL_PATH=/path/to/ggml-tiny.en.bin
make itest

# Check coverage
make coverage

# Run linter
golangci-lint run
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "Add MP3 decoder support

- Implement mp3Decoder using hajimehoshi/go-mp3
- Add unit tests for MP3 decoding
- Update documentation

Fixes #123"
```

**Commit Message Format**:
```
<type>: <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### 5. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Open pull request on GitHub
# Fill in the PR template with:
# - Description of changes
# - Related issues
# - Testing done
# - Checklist completion
```

## Code Standards

### Go Code Style

**Follow Go conventions**:
- Use `gofmt` for formatting (runs automatically via `make build`)
- Use `goimports` for import organization
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Keep functions small and focused

**Naming**:
```go
// Good
type TranscribeFile struct { ... }
func (uc *TranscribeFile) Execute(...)

// Bad
type transcribe_file struct { ... }
func (uc *transcribe_file) do_transcribe(...)
```

**Error Handling**:
```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("decode audio: %w", err)
}

// Bad: Generic error messages
if err != nil {
    return errors.New("error")
}
```

**Comments**:
```go
// Good: Explain "why" not "what"
// Resample to 16kHz because Whisper requires this sample rate
samples = resample(samples, 16000)

// Bad: Obvious comment
// Set sample rate to 16000
sampleRate = 16000
```

### Package Organization

**Hexagonal architecture layers**:
```
internal/
├── domain/        # Pure business entities (no dependencies)
├── port/          # Interfaces (contracts between layers)
├── usecase/       # Application logic (orchestration)
└── adapter/       # Implementations
    ├── inbound/   # Entry points (HTTP, CLI)
    └── outbound/  # External dependencies (Whisper, storage)
```

**Dependency rule**: Always point inward
```
Adapter → UseCase → Port → Domain
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed guidelines.

### Code Review Checklist

Before submitting PR, verify:

- [ ] Code follows Go conventions (`gofmt`, `goimports`)
- [ ] No linter warnings (`golangci-lint run`)
- [ ] All tests pass (`make test`)
- [ ] Coverage meets requirements (≥85% total, ≥90% usecase)
- [ ] Documentation updated (if adding features)
- [ ] No TODO comments (create issues instead)
- [ ] Error messages are descriptive
- [ ] Logging uses appropriate levels
- [ ] No hardcoded values (use config/env)

## Testing Requirements

### Test Coverage Gates

**Required Coverage** (enforced by CI):
- **Total**: ≥ 85%
- **UseCase package**: ≥ 90%

```bash
# Generate coverage report
make coverage

# View in browser
open coverage.html
```

### Writing Tests

**Unit Tests** (fast, isolated):
```go
func TestTranscribeFile_Execute(t *testing.T) {
    // Arrange: Create mocks
    mockRepo := &mockModelRepo{}
    mockTrans := &mockTranscriber{
        transcriptText: "expected result",
    }

    uc := &TranscribeFile{
        Repo: mockRepo,
        Trans: mockTrans,
        Factory: decoder.New,
    }

    // Act
    result, err := uc.Execute(context.Background(), TranscribeInput{
        Path: "testdata/test.wav",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "expected result", result.Text)
}
```

**Integration Tests** (slower, real dependencies):
```go
func TestMP3Decoder_Integration(t *testing.T) {
    if os.Getenv("GOSPER_INTEGRATION") == "" {
        t.Skip("set GOSPER_INTEGRATION=1 to run")
    }

    dec, err := decoder.NewMP3("testdata/test.mp3")
    require.NoError(t, err)
    defer dec.Close()

    samples, err := dec.DecodeAll()
    assert.NoError(t, err)
    assert.Greater(t, len(samples), 0)
}
```

**Test File Locations**:
- Unit tests: `*_test.go` next to implementation
- Integration tests: `test/integration/` or `*_test.go` with skip guard
- Test data: `testdata/` directory

**Table-Driven Tests**:
```go
func TestMP3Decoder_Validation(t *testing.T) {
    tests := []struct {
        name    string
        file    string
        wantErr string
    }{
        {"empty file", "testdata/empty.mp3", "empty file"},
        {"invalid format", "testdata/notmp3.txt", "invalid format"},
        {"oversized", "testdata/huge.mp3", "file too large"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := decoder.NewMP3(tt.file)
            assert.ErrorContains(t, err, tt.wantErr)
        })
    }
}
```

### Running Tests

```bash
# Unit tests only
make test

# With coverage
go test -cover ./...

# Specific package
go test ./internal/usecase/...

# Specific test
go test -run TestTranscribeFile_Execute ./internal/usecase/

# Integration tests
export GOSPER_INTEGRATION=1
export GOSPER_MODEL_PATH=/path/to/ggml-tiny.en.bin
make itest

# Verbose output
go test -v ./...

# Race detection
go test -race ./...
```

## Submitting Changes

### Pull Request Process

1. **Update Documentation**
   - Update README.md if adding user-facing features
   - Add/update docstrings for public APIs
   - Update relevant docs/ files

2. **Test Thoroughly**
   - Add tests for new functionality
   - Ensure all tests pass
   - Verify coverage requirements met

3. **Create Pull Request**
   - Use descriptive title
   - Fill out PR template completely
   - Link related issues
   - Add screenshots/examples if applicable

4. **Code Review**
   - Respond to feedback promptly
   - Make requested changes
   - Keep PR focused (avoid scope creep)

5. **Merge**
   - Maintainer will merge after approval
   - Delete your branch after merge

### Pull Request Template

```markdown
## Description
[Describe your changes]

## Related Issues
Fixes #123

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added and passing
- [ ] Coverage requirements met (≥85% total, ≥90% usecase)
```

## Architecture Guidelines

### Hexagonal Architecture Principles

**Key Rules**:
1. **Domain layer**: Pure Go types, no external dependencies
2. **Port layer**: Interfaces only, no implementations
3. **UseCase layer**: Orchestration logic, depends on Ports
4. **Adapter layer**: Implementations of Ports, external dependencies

**Adding a New Feature**:

#### Example: Add FLAC Format Support

**1. Domain Layer** (usually no changes):
```go
// No changes needed - Transcript and Segment remain the same
```

**2. Port Layer** (usually no changes):
```go
// No changes needed - Decoder interface already exists
type Decoder interface {
    DecodeAll() ([]float32, error)
    Info() Info
    Close() error
}
```

**3. UseCase Layer** (minimal or no changes):
```go
// No changes - TranscribeFile already uses DecoderFactory
```

**4. Adapter Layer** (new implementation):
```go
// internal/adapter/outbound/audio/decoder/flac.go
type flacDecoder struct { ... }

func NewFLAC(path string) (Decoder, error) {
    // Implement Decoder interface
}

// Update factory in decoder.go
func New(path string) (Decoder, error) {
    switch filepath.Ext(path) {
    case ".flac", ".FLAC":
        return NewFLAC(path)
    // ...
    }
}
```

**5. Tests**:
```go
// internal/adapter/outbound/audio/decoder/flac_test.go
func TestFLACDecoder_ValidFile(t *testing.T) { ... }

// internal/usecase/transcribe_file_test.go
func TestTranscribeFile_FLAC_Integration(t *testing.T) { ... }
```

### Dependency Injection

**Good** (inject dependencies):
```go
type TranscribeFile struct {
    Repo    port.ModelRepo
    Trans   port.Transcriber
    Factory DecoderFactory
}

func (uc *TranscribeFile) Execute(...) {
    dec, err := uc.Factory(input.Path)
    // ...
}
```

**Bad** (hard-coded dependencies):
```go
func (uc *TranscribeFile) Execute(...) {
    dec := decoder.NewWAV(input.Path)  // ❌ Tight coupling
    // ...
}
```

### Error Handling

**Wrap errors with context**:
```go
// Good
if err != nil {
    return fmt.Errorf("mp3: decode: %w", err)
}

// Bad
if err != nil {
    return err
}
```

**User-friendly error messages**:
```go
// Good
return fmt.Errorf("mp3: file too large (%d MB, max 200 MB)", size/1024/1024)

// Bad
return fmt.Errorf("file size exceeds limit")
```

## Documentation

### Code Documentation

**Package documentation**:
```go
// Package decoder provides audio format decoders for WAV and MP3 files.
//
// Decoders implement the Decoder interface and convert various audio formats
// to float32 mono PCM samples suitable for Whisper transcription.
package decoder
```

**Function documentation**:
```go
// NewMP3 creates a new MP3 decoder for the given file path.
//
// Returns an error if the file doesn't exist, is empty, too large (>200MB),
// or contains invalid MP3 data. The decoder must be closed when done to
// release file resources.
//
// Example:
//     dec, err := NewMP3("audio.mp3")
//     if err != nil {
//         return err
//     }
//     defer dec.Close()
//     samples, err := dec.DecodeAll()
func NewMP3(path string) (Decoder, error) {
    // ...
}
```

### User Documentation

When adding features, update:
- `README.md` - High-level overview
- `docs/QUICKSTART.md` - User examples
- `docs/API.md` - API endpoints
- `docs/CONFIGURATION.md` - Config options
- `docs/ARCHITECTURE.md` - Design decisions

## Common Contribution Types

### Bug Fixes

1. Create issue describing bug
2. Write failing test that reproduces bug
3. Fix bug
4. Verify test passes
5. Submit PR referencing issue

### New Features

1. Discuss in issue first (avoid wasted effort)
2. Design following hexagonal architecture
3. Implement with tests
4. Update documentation
5. Submit PR with examples

### Documentation

1. Identify gaps or outdated docs
2. Update or create documentation
3. Verify examples work
4. Submit PR

### Performance Improvements

1. Create benchmark showing issue
2. Profile to identify bottleneck
3. Implement optimization
4. Verify benchmark improves
5. Submit PR with before/after metrics

## Getting Help

**Questions about contributing?**
- Open a discussion on GitHub
- Check existing issues and PRs
- Read [ARCHITECTURE.md](ARCHITECTURE.md) for design patterns

**Stuck on something?**
- Ask in your PR for guidance
- Reference similar merged PRs
- Don't hesitate to ask for help!

## Recognition

Contributors will be:
- Listed in `CONTRIBUTORS.md`
- Credited in release notes
- Appreciated by the community!

Thank you for contributing to Gosper!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
