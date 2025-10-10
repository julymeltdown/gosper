package audio

import (
    "context"
    "fmt"
    "gosper/internal/domain"
    "gosper/internal/port"
)

type stubInput struct{}

func (stubInput) ListDevices(ctx context.Context) ([]domain.Device, error) { return nil, fmt.Errorf("audio input not built: build with -tags malgo") }
func (stubInput) Open(ctx context.Context, deviceID string, format domain.AudioFormat) (port.AudioStream, error) {
    return nil, fmt.Errorf("audio input not built: build with -tags malgo")
}
