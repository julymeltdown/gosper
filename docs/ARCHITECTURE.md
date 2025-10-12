# Architecture

Gosper is built using **Hexagonal Architecture** (also known as Ports and Adapters), which provides clean separation of concerns and makes the codebase highly testable and extensible.

## Why Hexagonal Architecture?

- **Testability**: Business logic can be tested without external dependencies
- **Flexibility**: Easily swap implementations (e.g., different audio decoders or storage backends)
- **Maintainability**: Clear boundaries between layers prevent coupling
- **Extensibility**: Add new adapters without modifying core logic

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                 Inbound Adapters                        │
│         (HTTP Server, CLI, Web UI)                      │
│  • cmd/server   - HTTP API entry point                 │
│  • cmd/gosper   - CLI interface                         │
│  • web/         - Static frontend                       │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                    Use Cases                            │
│          (Business Logic Orchestration)                 │
│  • TranscribeFile      - File → transcript pipeline     │
│  • RecordAndTranscribe - Mic → transcript pipeline      │
│  • ListDevices         - Audio device discovery         │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                      Ports                              │
│             (Interfaces / Contracts)                    │
│  • Transcriber   - Speech recognition interface         │
│  • Decoder       - Audio format decoder interface       │
│  • AudioInput    - Microphone capture interface         │
│  • ModelRepo     - Model management interface           │
│  • Storage       - File persistence interface           │
│  • Logger        - Logging interface                    │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                 Outbound Adapters                       │
│              (External Dependencies)                    │
│  • whisper/    - Whisper.cpp transcription             │
│  • decoder/    - WAV/MP3 audio decoders                │
│  • malgo/      - Microphone audio capture              │
│  • model/      - Model download & caching              │
│  • storage/    - File system operations                │
└─────────────────────────────────────────────────────────┘
```

## Layer Details

### Domain Layer (`internal/domain`)

Pure business entities with no external dependencies.

**Types**:
- `Transcript` - Transcription result with segments and metadata
- `Segment` - Individual speech segment with timing
- `ModelConfig` - Whisper model configuration

**Rules**:
- No dependencies on other layers
- Pure Go types, no external libraries
- Represents core business concepts

### Port Layer (`internal/port`)

Interfaces that define contracts between layers.

**Key Interfaces**:

```go
// Speech recognition
type Transcriber interface {
    Process(ctx context.Context, samples []float32, lang string) (domain.Transcript, error)
}

// Audio format decoding
type Decoder interface {
    DecodeAll() ([]float32, error)
    Info() Info
    Close() error
}

// Model management
type ModelRepo interface {
    Get(ctx context.Context, name string) (string, error)
}

// File storage
type Storage interface {
    WriteJSON(path string, data interface{}) error
    WriteText(path string, content string) error
}
```

**Purpose**:
- Define clear contracts for adapters
- Enable dependency injection for testing
- Document expected behavior

### UseCase Layer (`internal/usecase`)

Application logic that orchestrates business workflows.

**TranscribeFile UseCase** (`transcribe_file.go`):
```go
type TranscribeFile struct {
    Repo    port.ModelRepo      // Model management
    Trans   port.Transcriber    // Speech recognition
    Store   port.Storage        // File output
    Factory DecoderFactory      // Audio decoder factory
}

func (uc *TranscribeFile) Execute(ctx context.Context, in TranscribeInput) (domain.Transcript, error) {
    // 1. Decode audio file (via Factory → Decoder)
    // 2. Resample to 16kHz mono (Whisper requirement)
    // 3. Transcribe audio (via Transcriber)
    // 4. Save results (via Storage)
}
```

**Workflow**:
1. **Decode**: Use `DecoderFactory` to select decoder (WAV/MP3) by file extension
2. **Normalize**: Convert to float32, downmix stereo to mono
3. **Resample**: Linear resampling to 16kHz (Whisper requirement)
4. **Transcribe**: Process via Whisper model
5. **Store**: Save transcript as JSON/text

**RecordAndTranscribe UseCase** (`record_and_transcribe.go`):
```go
type RecordAndTranscribe struct {
    Repo  port.ModelRepo
    Trans port.Transcriber
    Input port.AudioInput     // Mic capture
    Store port.Storage
}
```

**Workflow**:
1. **Capture**: Record from microphone via `AudioInput`
2. **Buffer**: Accumulate audio samples in memory
3. **Transcribe**: Process via Whisper when recording stops
4. **Store**: Save transcript

### Adapter Layer (`internal/adapter`)

Concrete implementations of port interfaces.

#### Inbound Adapters

**HTTP Server** (`internal/adapter/inbound/http`):
- Exposes `/api/transcribe` endpoint
- Handles multipart file uploads
- Marshals responses to JSON
- Maps HTTP requests → UseCase calls

**CLI** (`internal/adapter/inbound/cli`):
- Cobra-based command structure
- `transcribe` command - file transcription
- `record` command - microphone capture
- `devices list` command - audio device discovery

#### Outbound Adapters

**Audio Decoders** (`internal/adapter/outbound/audio/decoder`):

**WAV Decoder** (`wav.go`):
- Supports PCM16 and Float32 formats
- Handles mono and stereo
- No file size limits
- Efficient streaming

**MP3 Decoder** (`mp3.go`):
- Uses `hajimehoshi/go-mp3` (pure Go, no CGO)
- Supports all MP3 bitrates (CBR, VBR, ABR)
- 200MB file size limit (memory protection)
- Always outputs stereo (downmixed to mono)

**Decoder Factory** (`decoder.go`):
```go
func New(path string) (Decoder, error) {
    ext := filepath.Ext(path)
    switch ext {
    case ".wav", ".Wave", ".WAV":
        return NewWAV(path)
    case ".mp3", ".MP3":
        return NewMP3(path)
    default:
        return nil, fmt.Errorf("unsupported format: %s", ext)
    }
}
```

**Whisper Transcriber** (`internal/adapter/outbound/whisper`):
- Wraps whisper.cpp Go bindings
- Manages model lifecycle
- Processes 16kHz mono float32 samples
- Returns segments with timestamps

**Model Repository** (`internal/adapter/outbound/model`):
- Downloads models from Hugging Face
- Caches models in OS cache directory
- Optional SHA256 verification
- Retry logic with exponential backoff

**Storage** (`internal/adapter/outbound/storage`):
- Atomic file writes (write to temp → rename)
- JSON marshaling for structured data
- Plain text output for transcripts

## Audio Processing Pipeline

### File Transcription Flow

```
┌──────────────┐
│  Audio File  │ (WAV/MP3, any sample rate, mono/stereo)
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│ DecoderFactory   │ Select decoder by file extension
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Decoder          │ Decode to float32 samples
│ (WAV or MP3)     │ - Normalize int16 → float32
└──────┬───────────┘ - Downmix stereo → mono
       │
       ▼
┌──────────────────┐
│ Resampler        │ Resample to 16kHz (Whisper requirement)
│ (Linear)         │ - Handles any source sample rate
└──────┬───────────┘ - Maintains audio duration
       │
       ▼
┌──────────────────┐
│ Whisper.cpp      │ Speech recognition
│ Transcriber      │ - Multi-language support
└──────┬───────────┘ - Automatic segmentation
       │
       ▼
┌──────────────────┐
│ Transcript       │ JSON/text output
└──────────────────┘
```

### Microphone Recording Flow

```
┌──────────────┐
│  Microphone  │ (System audio device)
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│ malgo Capture    │ Record float32 samples
│                  │ - Configurable device
└──────┬───────────┘ - Push-to-talk or toggle mode
       │
       ▼
┌──────────────────┐
│ Downmix          │ Stereo → mono (if needed)
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Buffer           │ Accumulate samples in memory
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Whisper.cpp      │ Transcribe when recording stops
│ Transcriber      │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ Transcript       │ Output to file or stdout
└──────────────────┘
```

## Dependency Direction

**Rule**: Dependencies always point inward (toward the domain).

```
Adapters → UseCases → Ports → Domain
```

**Examples**:
- ✅ UseCase depends on Port interface
- ✅ Adapter implements Port interface
- ✅ UseCase receives Adapter via dependency injection
- ❌ Port never depends on Adapter
- ❌ Domain never depends on UseCase
- ❌ UseCase never depends on specific Adapter (only interface)

## Extension Points

### Adding a New Audio Format

1. **Create decoder** in `internal/adapter/outbound/audio/decoder/`:
   ```go
   type flacDecoder struct { /* ... */ }

   func NewFLAC(path string) (Decoder, error) {
       // Implement Decoder interface
   }
   ```

2. **Update factory** in `decoder.go`:
   ```go
   case ".flac", ".FLAC":
       return NewFLAC(path)
   ```

3. **Write tests** in `flac_test.go`

**No changes needed** in UseCases or Ports!

### Adding a New Storage Backend

1. **Implement Port** interface:
   ```go
   type S3Storage struct { /* ... */ }

   func (s *S3Storage) WriteJSON(path string, data interface{}) error {
       // Upload to S3
   }
   ```

2. **Inject into UseCase**:
   ```go
   uc := &TranscribeFile{
       Store: &S3Storage{bucket: "transcripts"},
       // ...
   }
   ```

### Adding a New Inbound Interface

Example: gRPC API

1. **Create adapter** in `internal/adapter/inbound/grpc/`
2. **Define protobuf** service
3. **Map gRPC calls** → UseCase calls
4. **Inject dependencies** (UseCases with configured adapters)

## Testing Strategy

### Unit Tests

**UseCase Layer**:
- Mock all port interfaces
- Test business logic in isolation
- Example: `transcribe_file_test.go`

```go
func TestTranscribeFile_Execute(t *testing.T) {
    // Arrange: Create mocks
    mockRepo := &mockModelRepo{}
    mockTrans := &mockTranscriber{}
    mockStore := &mockStorage{}

    uc := &TranscribeFile{
        Repo: mockRepo,
        Trans: mockTrans,
        Store: mockStore,
    }

    // Act: Execute use case
    result, err := uc.Execute(ctx, input)

    // Assert: Verify behavior
    assert.NoError(t, err)
    assert.Equal(t, expectedText, result.Text)
}
```

**Adapter Layer**:
- Test concrete implementations
- Use real dependencies when lightweight (e.g., MP3 decoder)
- Example: `mp3_test.go`

```go
func TestMP3Decoder_ValidFile(t *testing.T) {
    dec, err := NewMP3("testdata/test.mp3")
    require.NoError(t, err)
    defer dec.Close()

    samples, err := dec.DecodeAll()
    assert.NoError(t, err)
    assert.Greater(t, len(samples), 0)
}
```

### Integration Tests

Test full pipeline with real implementations:

```go
func TestTranscribeFile_MP3_RealDecoder(t *testing.T) {
    // Use real MP3 decoder, mock Whisper (too slow for CI)
    uc := &TranscribeFile{
        Factory: decoder.New,  // Real factory
        Trans: &mockTranscriber{},  // Mock Whisper
        // ...
    }

    result, err := uc.Execute(ctx, TranscribeInput{
        Path: "testdata/test.mp3",
    })

    // Verify decoder → resampler → transcriber pipeline
}
```

### Test Coverage

Current gates enforced by CI:
- **Total coverage**: ≥ 85%
- **UseCase package**: ≥ 90%

Run tests:
```bash
make test        # Unit tests
make itest       # Integration tests (requires GOSPER_INTEGRATION=1)
make coverage    # Generate coverage report
```

## Build Tags

Gosper uses build tags to conditionally compile features:

- `cli` - Include CLI commands (`cmd/gosper`)
- `whisper` - Include Whisper inference (requires CGO)
- `malgo` - Include microphone capture (requires CGO)

**Examples**:
```bash
# CLI + Whisper transcription
go build -tags "cli whisper" -o gosper ./cmd/gosper

# CLI + Mic + Whisper (full featured)
go build -tags "cli malgo whisper" -o gosper ./cmd/gosper

# Server (HTTP API)
go build -tags "whisper" -o server ./cmd/server
```

**Why build tags?**
- Reduce binary size for specific use cases
- Avoid CGO dependency when not needed
- Separate concerns (CLI vs Server)

## Configuration Management

**Environment Variables** (see [CONFIGURATION.md](CONFIGURATION.md)):
- `GOSPER_MODEL` - Model name or path
- `GOSPER_LANG` - Language code or "auto"
- `GOSPER_THREADS` - Thread count for inference
- `GOSPER_CACHE` - Model cache directory

**Config File** (`~/.config/gosper/config.json`):
```json
{
  "LastDeviceID": "default",
  "AudioFeedback": true,
  "OutputDeviceID": "speakers",
  "BeepVolume": 0.5
}
```

Configuration is read at startup and passed to adapters via dependency injection.

## Design Decisions

### Why Factory Pattern for Decoders?

**Problem**: UseCase needs to select decoder based on file extension.

**Options**:
1. ❌ UseCase contains switch statement → couples UseCase to specific decoders
2. ✅ Factory function injected into UseCase → UseCase depends only on Decoder interface

**Benefit**: Can mock factory in tests, easy to add new formats.

### Why Linear Resampling?

**Trade-off**: Linear resampling is lower quality than sinc interpolation.

**Decision**: Acceptable for speech recognition (Whisper is robust to resampling artifacts).

**Future**: Add sinc resampler behind build tag for high-fidelity use cases.

### Why 200MB MP3 Limit?

**Problem**: MP3 decoding loads entire file into memory (~3x expansion).

**Risk**: Malicious users could upload huge files → OOM.

**Decision**: 200MB compressed = ~600MB decoded = reasonable limit.

**Alternative**: For larger files, recommend converting to WAV (no limit).

## Further Reading

- **[Build Guide](BUILD.md)** - Compile from source
- **[API Reference](API.md)** - HTTP API documentation
- **[Configuration](CONFIGURATION.md)** - Environment variables and settings
- **[Contributing](CONTRIBUTING.md)** - Development workflow and guidelines
