package port

import (
    "context"
    "io"
    "time"

    "gosper/internal/domain"
)

type Logger interface {
    Debug(ctx context.Context, msg string, kv ...any)
    Info(ctx context.Context, msg string, kv ...any)
    Warn(ctx context.Context, msg string, kv ...any)
    Error(ctx context.Context, msg string, kv ...any)
}

type Clock interface { Now() time.Time }

type ModelRepo interface {
    Ensure(ctx context.Context, modelName string) (localPath string, err error)
}

type Transcriber interface {
    // Accepts mono PCM @16kHz float32 samples.
    Transcribe(ctx context.Context, pcm16k []float32, cfg domain.ModelConfig) (domain.Transcript, error)
}

type AudioInput interface {
    ListDevices(ctx context.Context) ([]domain.Device, error)
    Open(ctx context.Context, deviceID string, fmt domain.AudioFormat) (AudioStream, error)
}

type AudioStream interface {
    Frames() <-chan []float32 // mono f32 frames
    Err() error
    Close() error
}

type Storage interface {
    WriteFile(ctx context.Context, path string, r io.Reader) error
    WriteTranscript(ctx context.Context, path string, t domain.Transcript) error
    TempPath(ctx context.Context, pattern string) (string, error)
}

