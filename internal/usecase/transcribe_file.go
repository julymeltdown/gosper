package usecase

import (
    "context"
    "fmt"
    "path/filepath"

    "gosper/internal/adapter/outbound/audio/decoder"
    "gosper/internal/adapter/outbound/audio/resample"
    "gosper/internal/domain"
    "gosper/internal/port"
    herr "gosper/pkg/errors"
)

// DecoderFactory allows tests to inject fake decoders.
type DecoderFactory func(path string) (decoder.Decoder, error)

type TranscribeFile struct {
    Repo    port.ModelRepo
    Trans   port.Transcriber
    Store   port.Storage
    Log     port.Logger
    Factory DecoderFactory
}

type TranscribeInput struct {
    Path          string
    OutPath       string // optional
    ModelName     string
    Language      string // default "auto"
    Translate     bool
    Threads       uint
    Timestamps    bool
    BeamSize      int
    MaxTokens     uint
    InitialPrompt string
}

func (uc *TranscribeFile) Execute(ctx context.Context, in TranscribeInput) (domain.Transcript, error) {
    if in.Path == "" {
        return domain.Transcript{}, herr.Wrap(herr.InvalidArgs, fmt.Errorf("missing input file path"))
    }
    if uc.Factory == nil {
        uc.Factory = decoder.New
    }
    dec, err := uc.Factory(in.Path)
    if err != nil {
        return domain.Transcript{}, herr.Wrap(herr.AudioError, err)
    }
    defer dec.Close()

    pcm, err := dec.DecodeAll()
    if err != nil {
        return domain.Transcript{}, herr.Wrap(herr.AudioError, err)
    }
    // resample to 16k mono
    pcm16k := resample.Linear(pcm, dec.Info().SampleRate, 16000)

    // model resolution
    modelPath, err := uc.Repo.Ensure(ctx, in.ModelName)
    if err != nil {
        return domain.Transcript{}, herr.Wrap(herr.ModelError, err)
    }

    cfg := domain.ModelConfig{
        ModelName:     filepath.Base(modelPath),
        ModelPath:     modelPath,
        Language:      orDefault(in.Language, "auto"),
        Translate:     in.Translate,
        Threads:       in.Threads,
        Timestamps:    in.Timestamps,
        BeamSize:      in.BeamSize,
        MaxTokens:     in.MaxTokens,
        InitialPrompt: in.InitialPrompt,
    }

    tr, err := uc.Transcribe(ctx, pcm16k, cfg)
    if err != nil {
        return domain.Transcript{}, herr.Wrap(herr.TranscriptionError, err)
    }

    if in.OutPath != "" {
        if err := uc.Store.WriteTranscript(ctx, in.OutPath, tr); err != nil {
            return tr, herr.Wrap(herr.FsError, err)
        }
    }
    return tr, nil
}

func (uc *TranscribeFile) Transcribe(ctx context.Context, pcm16k []float32, cfg domain.ModelConfig) (domain.Transcript, error) {
    return uc.Trans.Transcribe(ctx, pcm16k, cfg)
}

func orDefault[T comparable](v, def T) T {
    var zero T
    if v == zero {
        return def
    }
    return v
}

