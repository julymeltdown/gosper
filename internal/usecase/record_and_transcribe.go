package usecase

import (
    "context"
    "time"

    "gosper/internal/adapter/outbound/audio/resample"
    "gosper/internal/domain"
    "gosper/internal/port"
    herr "gosper/pkg/errors"
)

type RecordInput struct {
    DeviceID   string
    Duration   time.Duration // 0 => until context cancel
    ModelName  string
    Language   string
    Translate  bool
    Threads    uint
    Timestamps bool
    OutPath    string
}

type RecordAndTranscribe struct {
    Audio  port.AudioInput
    Repo   port.ModelRepo
    Trans  port.Transcriber
    Store  port.Storage
    Logger port.Logger
}

func (uc *RecordAndTranscribe) Execute(ctx context.Context, in RecordInput) (domain.Transcript, error) {
    fmt := domain.AudioFormat{SampleRate: 16000, Channels: 1, SampleType: "f32"}
    stream, err := uc.Audio.Open(ctx, in.DeviceID, fmt)
    if err != nil { return domain.Transcript{}, herr.Wrap(herr.AudioError, err) }
    defer stream.Close()

    // Capture frames until duration or cancel
    var buf []float32
    stop := make(chan struct{})
    if in.Duration > 0 {
        t := time.NewTimer(in.Duration)
        defer t.Stop()
        go func(){ <-t.C; close(stop) }()
    }
    frames := stream.Frames()
    loop:
    for {
        select {
        case <-ctx.Done():
            break loop
        case <-stop:
            break loop
        case fr, ok := <-frames:
            if !ok { break loop }
            buf = append(buf, fr...)
        }
    }
    if err := stream.Err(); err != nil { return domain.Transcript{}, herr.Wrap(herr.AudioError, err) }

    // buf is 16k mono already per contract; but resample anyway for safety
    pcm16k := resample.Linear(buf, fmt.SampleRate, 16000)

    // Resolve model
    modelPath, err := uc.Repo.Ensure(ctx, in.ModelName)
    if err != nil { return domain.Transcript{}, herr.Wrap(herr.ModelError, err) }

    cfg := domain.ModelConfig{ModelPath: modelPath, Language: in.Language, Translate: in.Translate, Threads: in.Threads, Timestamps: in.Timestamps}
    tr, err := uc.Trans.Transcribe(ctx, pcm16k, cfg)
    if err != nil { return domain.Transcript{}, herr.Wrap(herr.TranscriptionError, err) }
    if in.OutPath != "" {
        if err := uc.Store.WriteTranscript(ctx, in.OutPath, tr); err != nil { return tr, herr.Wrap(herr.FsError, err) }
    }
    return tr, nil
}

