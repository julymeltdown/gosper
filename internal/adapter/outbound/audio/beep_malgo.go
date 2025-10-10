//go:build malgo

package audio

import (
    "fmt"
    "math"
    "github.com/gen2brain/malgo"
    "time"
)

// PlayBeep plays a short 440Hz tone via default output device using malgo.
type BeepOptions struct {
    DeviceID string
    Volume   float32 // 0..1
    Freq     float64
    Duration time.Duration
}

func defaultBeepOptions() BeepOptions { return BeepOptions{Volume: 0.2, Freq: 440, Duration: 120 * time.Millisecond} }

func PlayBeep() { PlayBeepOptions(defaultBeepOptions()) }

func PlayBeepOptions(opts BeepOptions) {
    if opts.Duration <= 0 { opts.Duration = 120 * time.Millisecond }
    if opts.Freq <= 0 { opts.Freq = 440 }
    if opts.Volume <= 0 { opts.Volume = 0.2 }
    sr := 44100
    base := Tone(opts.Freq, opts.Duration, sr)
    pcm := make([]float32, len(base))
    for i := range base { pcm[i] = base[i] * opts.Volume }

    contextConfig := malgo.ContextConfig{}
    ctx, err := malgo.InitContext(nil, contextConfig, nil)
    if err != nil { ConsoleBell(); return }
    defer ctx.Uninit()

    cfg := malgo.DefaultDeviceConfig(malgo.Playback)
    cfg.Playback.Format = malgo.FormatF32
    cfg.Playback.Channels = 1
    cfg.SampleRate = uint32(sr)

    idx := 0
    callbacks := malgo.DeviceCallbacks{
        Data: func(pOutput, pInput []byte, frameCount uint32) {
            // write from pcm
            n := int(frameCount)
            for i := 0; i < n; i++ {
                var v float32
                if idx < len(pcm) { v = pcm[idx]; idx++ } else { v = 0 }
                // write little-endian float32
                u := math.Float32bits(v)
                off := i * 4
                if off+4 <= len(pOutput) {
                    pOutput[off] = byte(u)
                    pOutput[off+1] = byte(u >> 8)
                    pOutput[off+2] = byte(u >> 16)
                    pOutput[off+3] = byte(u >> 24)
                }
            }
        },
        Stop: func() { /* no-op */ },
    }

    // Optional output device selection by index:name
    if opts.DeviceID != "" {
        if devs, err := ctx.Devices(malgo.Playback); err == nil {
            // Accept index:N or name
            var idx int
            if _, e := fmt.Sscanf(opts.DeviceID, "index:%d", &idx); e == nil {
                if idx >= 0 && idx < len(devs) { cfg.Playback.DeviceID = devs[idx].ID.Pointer() }
            } else {
                for _, d := range devs { if d.Name() == opts.DeviceID { cfg.Playback.DeviceID = d.ID.Pointer(); break } }
            }
        }
    }

    dev, err := malgo.InitDevice(ctx.Context, cfg, callbacks)
    if err != nil { ConsoleBell(); return }
    defer dev.Uninit()
    if err := dev.Start(); err != nil { ConsoleBell(); return }
    // wait for tone duration
    time.Sleep(150 * time.Millisecond)
    _ = dev.Stop()
}
