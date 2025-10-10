package usecase

import (
    "context"
    "gosper/internal/domain"
    "gosper/internal/port"
)

type ListDevices struct { Audio port.AudioInput }

func (uc *ListDevices) Execute(ctx context.Context) ([]domain.Device, error) {
    return uc.Audio.ListDevices(ctx)
}

