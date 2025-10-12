//go:build malgo

package audio

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
	"gosper/internal/domain"
	"gosper/internal/port"
)

type MalgoInput struct{}

var _ port.AudioInput = (*MalgoInput)(nil)

func (MalgoInput) ListDevices(ctx context.Context) ([]domain.Device, error) {
	contextConfig := malgo.ContextConfig{}
	malgoCtx, err := malgo.InitContext(nil, contextConfig, nil)
	if err != nil {
		return nil, err
	}
	defer malgoCtx.Uninit()

	devs, err := malgoCtx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Device, 0, len(devs))
	for i, d := range devs {
		name := d.Name()
		out = append(out, domain.Device{ID: fmt.Sprintf("index:%d", i), Name: name})
	}
	return out, nil
}

func (MalgoInput) Open(ctx context.Context, deviceID string, format domain.AudioFormat) (port.AudioStream, error) {
	contextConfig := malgo.ContextConfig{}
	malgoCtx, err := malgo.InitContext(nil, contextConfig, nil)
	if err != nil {
		return nil, err
	}

	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = 1
	if format.Channels > 0 {
		cfg.Capture.Channels = uint32(format.Channels)
	}
	if format.SampleRate > 0 {
		cfg.SampleRate = uint32(format.SampleRate)
	} else {
		cfg.SampleRate = 16000
	}
	// Select device by index if provided
	if deviceID != "" {
		var idx int
		if _, err := fmt.Sscanf(deviceID, "index:%d", &idx); err == nil {
			devs, err := malgoCtx.Devices(malgo.Capture)
			if err == nil && idx >= 0 && idx < len(devs) {
				id := devs[idx].ID.Pointer()
				cfg.Capture.DeviceID = id
			}
		} else {
			// Try resolve by name (case-insensitive exact match)
			devs, err := malgoCtx.Devices(malgo.Capture)
			if err == nil {
				// Use the same resolver utilized by CLI
				// Map malgo devices to domain.Device
				list := make([]domain.Device, 0, len(devs))
				for i, d := range devs {
					list = append(list, domain.Device{ID: fmt.Sprintf("index:%d", i), Name: d.Name()})
				}
				id := ResolveDeviceID(list, deviceID)
				if id != "" {
					// parse back index:id
					var got int
					if _, err := fmt.Sscanf(id, "index:%d", &got); err == nil && got >= 0 && got < len(devs) {
						cfg.Capture.DeviceID = devs[got].ID.Pointer()
					}
				}
			}
		}
	}

	framesCh := make(chan []float32, 32)
	var dev *malgo.Device
	callbacks := malgo.DeviceCallbacks{
		Data: func(pOutput, pInput []byte, frameCount uint32) {
			// pInput contains interleaved float32 in little-endian format
			n := int(frameCount) * int(cfg.Capture.Channels)
			out := make([]float32, n)

			// Convert byte slice to float32 slice using encoding/binary for portability
			for i := 0; i < n && i*4+4 <= len(pInput); i++ {
				bits := binary.LittleEndian.Uint32(pInput[i*4 : i*4+4])
				out[i] = math.Float32frombits(bits)
			}
			// Downmix if channels > 1
			if cfg.Capture.Channels > 1 {
				mono := make([]float32, n/int(cfg.Capture.Channels))
				for f := 0; f < len(mono); f++ {
					sum := float32(0)
					base := f * int(cfg.Capture.Channels)
					for c := 0; c < int(cfg.Capture.Channels); c++ {
						sum += out[base+c]
					}
					mono[f] = sum / float32(cfg.Capture.Channels)
				}
				select {
				case framesCh <- mono:
				default:
				}
			} else {
				select {
				case framesCh <- out:
				default:
				}
			}
		},
	}

	dev, err = malgo.InitDevice(malgoCtx.Context, cfg, callbacks)
	if err != nil {
		malgoCtx.Uninit()
		return nil, err
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		malgoCtx.Uninit()
		return nil, err
	}

	s := &malgoStream{ctx: malgoCtx, dev: dev, ch: framesCh}
	go func() {
		<-ctx.Done()
		_ = s.Close()
	}()
	return s, nil
}

type malgoStream struct {
	mu  sync.Mutex
	ctx *malgo.AllocatedContext
	dev *malgo.Device
	ch  chan []float32
	err error
}

func (s *malgoStream) Frames() <-chan []float32 { return s.ch }
func (s *malgoStream) Err() error               { return s.err }

func (s *malgoStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dev != nil {
		_ = s.dev.Stop()
		s.dev.Uninit()
		s.dev = nil
	}
	if s.ctx != nil {
		s.ctx.Uninit()
		s.ctx = nil
	}
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
	return nil
}
