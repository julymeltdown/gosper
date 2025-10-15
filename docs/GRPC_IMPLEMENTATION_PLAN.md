# gRPC Implementation Plan for Gosper

## Executive Summary

Add gRPC support to Gosper as an alternative API interface alongside the existing HTTP REST API. This enables high-performance, type-safe communication with streaming support for large audio files.

## Architecture Analysis

### Current State
- **Protocol**: HTTP REST with multipart/form-data
- **Port**: 8080
- **Architecture**: Hexagonal (Ports & Adapters)
- **Layers**:
  - Domain: Business entities (`domain.Transcript`, `domain.ModelConfig`)
  - Port: Interfaces (`port.Transcriber`, `port.ModelRepo`)
  - UseCase: Application logic (`usecase.TranscribeFile`)
  - Adapter: HTTP handler in `cmd/server/main.go`

### Proposed gRPC Integration

```
┌─────────────────────────────────────────────────┐
│                   Clients                        │
│   HTTP (8080)              gRPC (50051)          │
└────────┬──────────────────────┬─────────────────┘
         │                      │
         ▼                      ▼
┌─────────────────┐    ┌──────────────────┐
│  HTTP Handler   │    │  gRPC Handler    │
│  (existing)     │    │  (NEW)           │
└────────┬────────┘    └────────┬─────────┘
         │                      │
         └──────────┬───────────┘
                    ▼
         ┌─────────────────────┐
         │ UseCase Layer       │
         │ TranscribeFile      │
         └─────────────────────┘
                    │
         ┌──────────┴──────────┐
         ▼                     ▼
   ┌─────────┐          ┌─────────┐
   │Whisper  │          │Model    │
   │Adapter  │          │Repo     │
   └─────────┘          └─────────┘
```

**Key Decision**: Run both HTTP and gRPC on separate ports in the same process, sharing the usecase layer.

## Implementation Plan

### Phase 1: Protocol Buffer Definition

#### 1.1 File Structure
```
gosper/
├── api/
│   └── proto/
│       └── gosper/
│           └── v1/
│               ├── transcription.proto
│               └── buf.yaml (optional: if using buf)
├── pkg/
│   └── grpc/
│       └── gen/
│           └── go/
│               └── gosper/
│                   └── v1/
│                       ├── transcription.pb.go
│                       └── transcription_grpc.pb.go
```

#### 1.2 Service Definition (`api/proto/gosper/v1/transcription.proto`)

```protobuf
syntax = "proto3";

package gosper.v1;

option go_package = "gosper/pkg/grpc/gen/go/gosper/v1;gosperv1";

// TranscriptionService provides audio-to-text transcription
service TranscriptionService {
  // Transcribe converts audio to text using client streaming
  // Client sends audio chunks, server returns final transcript
  rpc Transcribe(stream TranscribeRequest) returns (TranscribeResponse);

  // TranscribeWithProgress provides real-time progress updates
  // Bidirectional streaming for progress tracking
  rpc TranscribeWithProgress(stream TranscribeRequest) returns (stream TranscribeProgressResponse);

  // Health check
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

// Audio chunk or config message
message TranscribeRequest {
  oneof data {
    TranscribeConfig config = 1;  // First message must be config
    AudioChunk audio_chunk = 2;    // Subsequent messages are audio
  }
}

// Transcription configuration
message TranscribeConfig {
  string model = 1;                 // Model name (e.g., "ggml-tiny.en.bin")
  string language = 2;              // Language code or "auto"
  bool translate = 3;               // Translate to English
  uint32 threads = 4;               // Number of threads
  bool timestamps = 5;              // Include timestamps
  int32 beam_size = 6;              // Beam search size
  uint32 max_tokens = 7;            // Max tokens per segment
  string initial_prompt = 8;        // Context for transcription
  AudioFormat format = 9;           // Audio format
}

// Audio format metadata
message AudioFormat {
  string encoding = 1;              // "wav", "mp3", "webm", "ogg"
  int32 sample_rate = 2;            // Original sample rate
  int32 channels = 3;               // Number of channels
}

// Audio data chunk
message AudioChunk {
  bytes data = 1;                   // Raw audio bytes
  int64 sequence_number = 2;        // Chunk sequence for ordering
}

// Final transcription result
message TranscribeResponse {
  string language = 1;              // Detected language
  string text = 2;                  // Full transcription text
  repeated Segment segments = 3;    // Time-stamped segments
  int64 duration_ms = 4;            // Processing duration
}

// Progress update (for streaming mode)
message TranscribeProgressResponse {
  oneof event {
    ProgressUpdate progress = 1;
    TranscribeResponse result = 2;
  }
}

message ProgressUpdate {
  float percent_complete = 1;       // 0.0 to 1.0
  string status = 2;                // Human-readable status
  int64 elapsed_ms = 3;             // Elapsed time
}

// Time-stamped segment
message Segment {
  int32 index = 1;                  // Segment number
  int64 start_ms = 2;               // Start time in milliseconds
  int64 end_ms = 3;                 // End time in milliseconds
  string text = 4;                  // Segment text
}

// Health check request/response
message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;                // "ok" or error message
  string version = 2;               // Server version
}
```

**Design Rationale**:
- **Client streaming**: Efficient for large audio files (avoids loading entire file in memory)
- **Bidirectional streaming option**: Real-time progress for long transcriptions
- **oneof pattern**: Clean config-then-data flow
- **AudioFormat**: Allows server-side format conversion (ffmpeg)
- **Sequence numbers**: Ensures correct chunk ordering

### Phase 2: Tooling Setup

#### 2.1 Install Tools

Add to `tools.go` (for `go install` tracking):
```go
//go:build tools
// +build tools

package tools

import (
    _ "google.golang.org/protobuf/cmd/protoc-gen-go"
    _ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)
```

#### 2.2 Update `go.mod`
```bash
go get google.golang.org/grpc@latest
go get google.golang.org/protobuf@latest
```

#### 2.3 Create Makefile Target
```makefile
# Generate protobuf code
.PHONY: proto
proto:
	mkdir -p pkg/grpc/gen/go
	protoc \
		--go_out=pkg/grpc/gen/go \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg/grpc/gen/go \
		--go-grpc_opt=paths=source_relative \
		--proto_path=api/proto \
		api/proto/gosper/v1/*.proto
```

**Alternative: Use `buf`** (recommended for teams)
```yaml
# buf.yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

### Phase 3: gRPC Server Implementation

#### 3.1 Create gRPC Adapter (`internal/adapter/inbound/grpc/server.go`)

```go
package grpc

import (
    "context"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    pb "gosper/pkg/grpc/gen/go/gosper/v1"
    "gosper/internal/domain"
    "gosper/internal/usecase"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type Server struct {
    pb.UnimplementedTranscriptionServiceServer
    transcribeUseCase *usecase.TranscribeFile
    logger            Logger
}

type Logger interface {
    Printf(format string, v ...interface{})
}

func NewServer(uc *usecase.TranscribeFile, logger Logger) *Server {
    return &Server{
        transcribeUseCase: uc,
        logger:            logger,
    }
}

// Transcribe handles client-streaming audio upload
func (s *Server) Transcribe(stream pb.TranscriptionService_TranscribeServer) error {
    // 1. Receive config (first message)
    req, err := stream.Recv()
    if err != nil {
        return status.Error(codes.InvalidArgument, "failed to receive config")
    }

    config, ok := req.GetData().(*pb.TranscribeRequest_Config)
    if !ok {
        return status.Error(codes.InvalidArgument, "first message must be config")
    }

    // 2. Create temp file for audio assembly
    tmpDir := os.TempDir()
    ext := extensionFromFormat(config.Config.GetFormat().GetEncoding())
    tmpFile, err := os.CreateTemp(tmpDir, "grpc-upload-*"+ext)
    if err != nil {
        return status.Error(codes.Internal, fmt.Sprintf("create temp: %v", err))
    }
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()

    // 3. Receive audio chunks and write to temp file
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return status.Error(codes.Internal, fmt.Sprintf("receive chunk: %v", err))
        }

        chunk, ok := req.GetData().(*pb.TranscribeRequest_AudioChunk)
        if !ok {
            return status.Error(codes.InvalidArgument, "expected audio chunk")
        }

        if _, err := tmpFile.Write(chunk.AudioChunk.GetData()); err != nil {
            return status.Error(codes.Internal, fmt.Sprintf("write chunk: %v", err))
        }
    }

    // Ensure all data is written
    if err := tmpFile.Sync(); err != nil {
        return status.Error(codes.Internal, fmt.Sprintf("sync file: %v", err))
    }

    // 4. Call transcription use case
    start := time.Now()
    result, err := s.transcribeUseCase.Execute(stream.Context(), usecase.TranscribeInput{
        Path:          tmpFile.Name(),
        ModelName:     orDefault(config.Config.GetModel(), "ggml-tiny.en.bin"),
        Language:      orDefault(config.Config.GetLanguage(), "auto"),
        Translate:     config.Config.GetTranslate(),
        Threads:       uint(config.Config.GetThreads()),
        Timestamps:    config.Config.GetTimestamps(),
        BeamSize:      int(config.Config.GetBeamSize()),
        MaxTokens:     uint(config.Config.GetMaxTokens()),
        InitialPrompt: config.Config.GetInitialPrompt(),
    })
    if err != nil {
        s.logger.Printf("transcription error: %v", err)
        return status.Error(codes.Internal, "transcription failed")
    }

    // 5. Convert domain result to protobuf
    pbSegments := make([]*pb.Segment, len(result.Segments))
    for i, seg := range result.Segments {
        pbSegments[i] = &pb.Segment{
            Index:   int32(seg.Index),
            StartMs: seg.StartMS,
            EndMs:   seg.EndMS,
            Text:    seg.Text,
        }
    }

    // 6. Send response
    response := &pb.TranscribeResponse{
        Language:   result.Language,
        Text:       result.FullText,
        Segments:   pbSegments,
        DurationMs: time.Since(start).Milliseconds(),
    }

    return stream.SendAndClose(response)
}

// HealthCheck implements health checking
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
    return &pb.HealthCheckResponse{
        Status:  "ok",
        Version: "1.0.0", // TODO: Read from build info
    }, nil
}

func extensionFromFormat(encoding string) string {
    switch encoding {
    case "wav":
        return ".wav"
    case "mp3":
        return ".mp3"
    case "webm":
        return ".webm"
    case "ogg":
        return ".ogg"
    default:
        return ".wav"
    }
}

func orDefault(v, def string) string {
    if v == "" {
        return def
    }
    return v
}
```

#### 3.2 Update `cmd/server/main.go`

```go
package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "gosper/internal/adapter/inbound/grpc"
    "gosper/internal/adapter/outbound/model"
    "gosper/internal/adapter/outbound/storage"
    "gosper/internal/adapter/outbound/whispercpp"
    "gosper/internal/config"
    "gosper/internal/usecase"

    pb "gosper/pkg/grpc/gen/go/gosper/v1"
    grpcLib "google.golang.org/grpc"
)

func main() {
    cfg := config.FromEnv()
    logger := log.New(os.Stdout, "", log.LstdFlags)

    // Shared use case
    transcribeUC := &usecase.TranscribeFile{
        Repo:  &model.FSRepo{BaseURL: cfg.ModelBaseURL},
        Trans: &whispercpp.Transcriber{},
        Store: storage.FS{},
    }

    // HTTP Server (existing)
    app := &application{
        cfg:    cfg,
        logger: logger,
        usecases: struct{ transcribeFile *usecase.TranscribeFile }{
            transcribeFile: transcribeUC,
        },
    }
    httpSrv := NewServer(app)

    // gRPC Server (new)
    grpcServer := grpcLib.NewServer()
    pb.RegisterTranscriptionServiceServer(grpcServer, grpc.NewServer(transcribeUC, logger))

    // Start HTTP server
    go func() {
        logger.Printf("HTTP server listening on %s", cfg.Addr)
        if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    // Start gRPC server
    grpcPort := os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "50051"
    }
    grpcListener, err := net.Listen("tcp", ":"+grpcPort)
    if err != nil {
        log.Fatalf("failed to listen on gRPC port: %v", err)
    }
    go func() {
        logger.Printf("gRPC server listening on :%s", grpcPort)
        if err := grpcServer.Serve(grpcListener); err != nil {
            log.Fatalf("gRPC server error: %v", err)
        }
    }()

    // Graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    logger.Println("Shutting down servers...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    grpcServer.GracefulStop()
    if err := httpSrv.Shutdown(ctx); err != nil {
        log.Fatalf("shutdown error: %v", err)
    }
}

// ... rest of existing code (HTTP handlers, etc.)
```

### Phase 4: Deployment Configuration

#### 4.1 Update `Dockerfile.server`

```dockerfile
# Expose both HTTP and gRPC ports
EXPOSE 8080
EXPOSE 50051

ENV PORT=8080
ENV GRPC_PORT=50051
```

#### 4.2 Update `deploy/k8s/base/backend.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gosper-be
  namespace: ${NAMESPACE}
spec:
  replicas: 1
  selector:
    matchLabels: { app: gosper-be }
  template:
    metadata:
      labels: { app: gosper-be }
    spec:
      containers:
        - name: be
          image: ${BE_IMAGE}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080  # HTTP
              name: http
            - containerPort: 50051  # gRPC
              name: grpc
          env:
            - name: PORT
              value: "8080"
            - name: GRPC_PORT
              value: "50051"
          resources:
            requests: { cpu: "500m", memory: "512Mi" }
            limits: { cpu: "2000m", memory: "2Gi" }
---
apiVersion: v1
kind: Service
metadata:
  name: gosper-be
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  selector: { app: gosper-be }
  ports:
    - name: http
      port: 80
      targetPort: 8080
      nodePort: ${BE_HTTP_NODEPORT}
    - name: grpc
      port: 50051
      targetPort: 50051
      nodePort: ${BE_GRPC_NODEPORT}
```

### Phase 5: Client Examples

#### 5.1 Go Client (`examples/grpc/go/client.go`)

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "time"

    pb "gosper/pkg/grpc/gen/go/gosper/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    if len(os.Args) < 2 {
        log.Fatal("Usage: client <audio-file>")
    }

    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewTranscriptionServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    stream, err := client.Transcribe(ctx)
    if err != nil {
        log.Fatalf("transcribe: %v", err)
    }

    // Send config
    err = stream.Send(&pb.TranscribeRequest{
        Data: &pb.TranscribeRequest_Config{
            Config: &pb.TranscribeConfig{
                Model:    "ggml-tiny.en.bin",
                Language: "auto",
                Format: &pb.AudioFormat{
                    Encoding: "mp3",
                },
            },
        },
    })
    if err != nil {
        log.Fatalf("send config: %v", err)
    }

    // Send audio chunks
    file, err := os.Open(os.Args[1])
    if err != nil {
        log.Fatalf("open file: %v", err)
    }
    defer file.Close()

    buf := make([]byte, 64*1024) // 64KB chunks
    seq := int64(0)
    for {
        n, err := file.Read(buf)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("read: %v", err)
        }

        err = stream.Send(&pb.TranscribeRequest{
            Data: &pb.TranscribeRequest_AudioChunk{
                AudioChunk: &pb.AudioChunk{
                    Data:           buf[:n],
                    SequenceNumber: seq,
                },
            },
        })
        if err != nil {
            log.Fatalf("send chunk: %v", err)
        }
        seq++
    }

    // Receive result
    resp, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatalf("receive: %v", err)
    }

    fmt.Printf("Language: %s\n", resp.Language)
    fmt.Printf("Text: %s\n", resp.Text)
    fmt.Printf("Duration: %dms\n", resp.DurationMs)
    fmt.Printf("Segments: %d\n", len(resp.Segments))
}
```

#### 5.2 Python Client (`examples/grpc/python/client.py`)

```python
#!/usr/bin/env python3
import sys
import grpc
from gosper.v1 import transcription_pb2, transcription_pb2_grpc

def transcribe(audio_file: str, server: str = "localhost:50051"):
    with grpc.insecure_channel(server) as channel:
        stub = transcription_pb2_grpc.TranscriptionServiceStub(channel)

        def request_iterator():
            # Send config first
            yield transcription_pb2.TranscribeRequest(
                config=transcription_pb2.TranscribeConfig(
                    model="ggml-tiny.en.bin",
                    language="auto",
                    format=transcription_pb2.AudioFormat(encoding="mp3")
                )
            )

            # Send audio chunks
            with open(audio_file, "rb") as f:
                seq = 0
                while True:
                    chunk = f.read(64 * 1024)  # 64KB
                    if not chunk:
                        break
                    yield transcription_pb2.TranscribeRequest(
                        audio_chunk=transcription_pb2.AudioChunk(
                            data=chunk,
                            sequence_number=seq
                        )
                    )
                    seq += 1

        response = stub.Transcribe(request_iterator())

        print(f"Language: {response.language}")
        print(f"Text: {response.text}")
        print(f"Duration: {response.duration_ms}ms")
        print(f"Segments: {len(response.segments)}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: client.py <audio-file>")
        sys.exit(1)
    transcribe(sys.argv[1])
```

### Phase 6: Testing

#### 6.1 Integration Test (`test/grpc/transcribe_test.go`)

```go
package grpc_test

import (
    "context"
    "io"
    "net"
    "os"
    "testing"
    "time"

    "gosper/internal/adapter/inbound/grpc"
    "gosper/internal/usecase"
    pb "gosper/pkg/grpc/gen/go/gosper/v1"

    grpcLib "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/test/bufconn"
)

func TestTranscribe(t *testing.T) {
    // Setup
    listener := bufconn.Listen(1024 * 1024)
    s := grpcLib.NewServer()

    // Mock use case
    mockUC := &usecase.TranscribeFile{
        // ... mock dependencies
    }

    pb.RegisterTranscriptionServiceServer(s, grpc.NewServer(mockUC, testLogger{}))

    go func() {
        if err := s.Serve(listener); err != nil {
            t.Errorf("serve: %v", err)
        }
    }()
    defer s.Stop()

    // Connect
    ctx := context.Background()
    conn, err := grpcLib.DialContext(ctx, "",
        grpcLib.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return listener.Dial()
        }),
        grpcLib.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        t.Fatalf("dial: %v", err)
    }
    defer conn.Close()

    client := pb.NewTranscriptionServiceClient(conn)

    // Test
    stream, err := client.Transcribe(ctx)
    if err != nil {
        t.Fatalf("transcribe: %v", err)
    }

    // Send test data
    // ... (implementation)
}

type testLogger struct{}
func (testLogger) Printf(format string, v ...interface{}) {}
```

## Migration Strategy

### Backward Compatibility
1. **Keep HTTP API unchanged** - Existing clients continue working
2. **gRPC as opt-in** - Clients choose to migrate when ready
3. **Shared business logic** - Both APIs use same usecase layer

### Rollout Plan
1. **Week 1**: Implement proto and generate code
2. **Week 2**: Implement gRPC server and basic tests
3. **Week 3**: Deploy to staging, write client examples
4. **Week 4**: Production deployment with monitoring
5. **Week 5+**: Document and promote gRPC to users

### Monitoring
- Add gRPC metrics (requests, latency, errors)
- Log both HTTP and gRPC usage
- Track adoption rate

## Performance Considerations

### Chunking Strategy
- **Chunk size**: 64KB (balance between overhead and memory)
- **Memory efficiency**: Stream directly to disk, don't buffer entire file
- **Backpressure**: gRPC handles flow control automatically

### Concurrency
- Each transcription spawns goroutines for:
  - Audio chunk receiving
  - Whisper processing (CPU-bound)
- Limit concurrent transcriptions to prevent OOM

### Optimization Opportunities
1. **Connection pooling**: Reuse HTTP client for model downloads
2. **Model caching**: Keep loaded models in memory (future work)
3. **Batch processing**: Group small transcriptions (future work)

## Security

### Authentication (Future)
```protobuf
// Add to service
rpc Authenticate(AuthRequest) returns (AuthResponse);

message AuthRequest {
  string api_key = 1;
}
```

### TLS (Production)
```go
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
grpcServer := grpc.NewServer(grpc.Creds(creds))
```

## Documentation Deliverables

1. **API docs**: Update docs/API.md with gRPC section
2. **Client examples**: Go, Python, JavaScript/TypeScript
3. **Migration guide**: HTTP → gRPC transition
4. **Proto documentation**: Auto-generate from proto comments

## Success Criteria

- [ ] Proto file compiled without errors
- [ ] gRPC server runs alongside HTTP
- [ ] Integration tests pass (>90% coverage)
- [ ] Client examples work in 3+ languages
- [ ] No regression in HTTP API performance
- [ ] gRPC latency < HTTP for files >10MB
- [ ] Deployed to production without incidents

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes to HTTP API | HIGH | Keep HTTP completely separate |
| Port conflicts in deployment | MEDIUM | Use env vars for port config |
| Client compatibility issues | MEDIUM | Provide multiple client examples |
| Increased memory usage | MEDIUM | Implement streaming + limits |
| Proto versioning complexity | LOW | Use v1 namespace, plan v2 early |

## Timeline Estimate

- **Proto definition**: 1 day
- **Code generation setup**: 0.5 days
- **gRPC server implementation**: 2 days
- **Testing**: 1.5 days
- **Deployment updates**: 1 day
- **Client examples**: 1 day
- **Documentation**: 1 day

**Total**: ~8 days (1.5 sprint iterations)

## Next Steps

1. Review this plan with team
2. Set up proto file in `api/proto/`
3. Install protobuf tools
4. Generate initial code
5. Implement basic gRPC server
6. Write first integration test
