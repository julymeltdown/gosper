package usecase

import (
    "context"
    "errors"
    "io"
    "testing"
    "time"

    "gosper/internal/domain"
    "gosper/internal/port"
    
)

// fakes for AudioInput and stream
type fakeStream struct{ ch chan []float32; err error }
func (s *fakeStream) Frames() <-chan []float32 { return s.ch }
func (s *fakeStream) Err() error { return s.err }
func (s *fakeStream) Close() error { close(s.ch); return nil }

type fakeAudio struct{ stream *fakeStream }
func (a *fakeAudio) ListDevices(ctx context.Context) ([]domain.Device, error) { return nil, nil }
func (a *fakeAudio) Open(ctx context.Context, deviceID string, fmt domain.AudioFormat) (port.AudioStream, error) { return a.stream, nil }

type fakeRepo2 struct{ path string; err error }
func (f *fakeRepo2) Ensure(ctx context.Context, name string) (string, error) { return f.path, f.err }

type fakeTranscriber2 struct{ called bool; err error }
func (t *fakeTranscriber2) Transcribe(ctx context.Context, pcm []float32, cfg domain.ModelConfig) (domain.Transcript, error) {
    t.called = true
    if t.err != nil { return domain.Transcript{}, t.err }
    return domain.Transcript{FullText: "ok"}, nil
}

type fakeStorage2 struct{}
func (s *fakeStorage2) WriteFile(ctx context.Context, path string, r io.Reader) error { return nil }
func (s *fakeStorage2) WriteTranscript(ctx context.Context, path string, t domain.Transcript) error { return nil }
func (s *fakeStorage2) TempPath(ctx context.Context, pattern string) (string, error) { return "/tmp/x", nil }

type errAudio struct{}
func (errAudio) ListDevices(ctx context.Context) ([]domain.Device, error) { return nil, nil }
func (errAudio) Open(ctx context.Context, deviceID string, fmt domain.AudioFormat) (port.AudioStream, error) { return nil, errors.New("no device") }

func TestRecordAndTranscribe_DurationStopsAndTranscribes(t *testing.T) {
    frames := make(chan []float32, 10)
    // push some frames
    frames <- make([]float32, 1600)
    frames <- make([]float32, 1600)
    stream := &fakeStream{ ch: frames }
    audio := &fakeAudio{ stream: stream }
    repo := &fakeRepo2{ path: "/m" }
    tr := &fakeTranscriber2{}
    uc := &RecordAndTranscribe{ Audio: audio, Repo: repo, Trans: tr, Store: &fakeStorage2{} }

    ctx := context.Background()
    _, err := uc.Execute(ctx, RecordInput{ Duration: 50 * time.Millisecond, ModelName: "/m" })
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if !tr.called { t.Fatalf("expected transcriber to be called") }
}

func TestRecordAndTranscribe_PropagatesErrors(t *testing.T) {
    // audio open error
    uc1 := &RecordAndTranscribe{ Audio: errAudio{} }
    if _, err := uc1.Execute(context.Background(), RecordInput{}); err == nil { t.Fatal("expected error") }

    // transcriber error
    frames := make(chan []float32, 1)
    frames <- make([]float32, 1600)
    stream := &fakeStream{ ch: frames }
    audio := &fakeAudio{ stream: stream }
    repo := &fakeRepo2{ path: "/m" }
    tr := &fakeTranscriber2{ err: errors.New("boom") }
    uc2 := &RecordAndTranscribe{ Audio: audio, Repo: repo, Trans: tr, Store: &fakeStorage2{} }
    if _, err := uc2.Execute(context.Background(), RecordInput{ Duration: 1 * time.Millisecond, ModelName: "/m" }); err == nil { t.Fatal("expected error") }
}
