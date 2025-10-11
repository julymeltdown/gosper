//go:build whisper

package whispercpp

import (
    "context"
    w "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
    "gosper/internal/domain"
    "gosper/internal/port"
)

type Transcriber struct{}

var _ port.Transcriber = (*Transcriber)(nil)

func (t *Transcriber) Transcribe(ctx context.Context, pcm16k []float32, cfg domain.ModelConfig) (domain.Transcript, error) {
    model, err := w.New(cfg.ModelPath)
    if err != nil { return domain.Transcript{}, err }
    defer model.Close()

    c, err := model.NewContext()
    if err != nil { return domain.Transcript{}, err }
    if cfg.Language != "" { _ = c.SetLanguage(cfg.Language) }
    c.SetTranslate(cfg.Translate)
    if cfg.Threads > 0 { c.SetThreads(cfg.Threads) }
    c.SetTokenTimestamps(cfg.Timestamps)
    if cfg.BeamSize > 0 { c.SetBeamSize(cfg.BeamSize) }
    if cfg.MaxTokens > 0 { c.SetMaxTokensPerSegment(cfg.MaxTokens) }
    if cfg.InitialPrompt != "" { c.SetInitialPrompt(cfg.InitialPrompt) }

    var segments []domain.TranscriptSegment
    segCB := func(s w.Segment) {
        segments = append(segments, domain.TranscriptSegment{
            Index:   s.Num,
            StartMS: int64(s.Start / 1e6),
            EndMS:   int64(s.End / 1e6),
            Text:    s.Text,
        })
    }
    if err := c.Process(pcm16k, nil, segCB, nil); err != nil { return domain.Transcript{}, err }
    // Drain any remaining
    for {
        seg, err := c.NextSegment()
        if err != nil { break }
        segments = append(segments, domain.TranscriptSegment{
            Index:   seg.Num,
            StartMS: int64(seg.Start / 1e6),
            EndMS:   int64(seg.End / 1e6),
            Text:    seg.Text,
        })
    }
    var full string
    for _, s := range segments { if full == "" { full = s.Text } else { full += s.Text } }
    tr := domain.Transcript{Language: cfg.Language, Segments: segments, FullText: full}
    _ = ctx // reserved for future cancellation integration
    return tr, nil
}

