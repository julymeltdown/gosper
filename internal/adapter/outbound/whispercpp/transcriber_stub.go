package whispercpp

import (
    "context"
    "fmt"
    "gosper/internal/domain"
    "gosper/internal/port"
)

type Transcriber struct{}

var _ port.Transcriber = (*Transcriber)(nil)

func (t *Transcriber) Transcribe(ctx context.Context, pcm16k []float32, cfg domain.ModelConfig) (domain.Transcript, error) {
    return domain.Transcript{}, fmt.Errorf("whisper adapter not built: build with -tags whisper")
}

