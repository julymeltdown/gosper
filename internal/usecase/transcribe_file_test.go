package usecase

import (
    "context"
    "errors"
    "io"
    "os"
    "testing"

    "gosper/internal/adapter/outbound/audio/decoder"
    "gosper/internal/domain"
)

// Fake implementations
type fakeDecoder struct{ sr, ch int; pcm []float32 }
func (f *fakeDecoder) Info() decoder.Info { return decoder.Info{SampleRate: f.sr, Channels: f.ch, Frames: int64(len(f.pcm))} }
func (f *fakeDecoder) DecodeAll() ([]float32, error) { return f.pcm, nil }
func (f *fakeDecoder) Close() error { return nil }

type fakeRepo struct{ path string; err error }
func (f *fakeRepo) Ensure(ctx context.Context, name string) (string, error) { return f.path, f.err }

type fakeTranscriber struct{ gotPCM int; cfg domain.ModelConfig; err error }
func (t *fakeTranscriber) Transcribe(ctx context.Context, pcm []float32, cfg domain.ModelConfig) (domain.Transcript, error) {
    t.gotPCM = len(pcm)
    t.cfg = cfg
    if t.err != nil { return domain.Transcript{}, t.err }
    return domain.Transcript{Language: cfg.Language, FullText: "hello world"}, nil
}

type fakeStorage struct{ wrote bool; path string; last domain.Transcript; err error }
func (s *fakeStorage) WriteFile(ctx context.Context, path string, r io.Reader) error { return nil }
func (s *fakeStorage) WriteTranscript(ctx context.Context, path string, t domain.Transcript) error { s.wrote=true; s.path=path; s.last=t; return s.err }
func (s *fakeStorage) TempPath(ctx context.Context, pattern string) (string, error) { return "/tmp/x", nil }

func TestTranscribeFile_HappyPath(t *testing.T) {
    // 8kHz mono 1-second constant signal; after resample expect ~16000 samples
    inPCM := make([]float32, 8000)
    for i := range inPCM { inPCM[i] = 0.5 }
    dec := &fakeDecoder{sr: 8000, ch: 1, pcm: inPCM}
    rep := &fakeRepo{path: "/models/ggml-tiny.en.bin"}
    tr := &fakeTranscriber{}
    st := &fakeStorage{}
    uc := &TranscribeFile{
        Repo: rep, Trans: tr, Store: st, Factory: func(string)(decoder.Decoder,error){ return dec, nil },
    }
    got, err := uc.Execute(context.Background(), TranscribeInput{
        Path: "dummy.wav", OutPath: "out.txt", ModelName: rep.path, Language: "en",
    })
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if got.FullText != "hello world" { t.Fatalf("unexpected transcript: %q", got.FullText) }
    if tr.gotPCM < 15900 || tr.gotPCM > 16100 { t.Fatalf("unexpected resampled len: %d", tr.gotPCM) }
    if !st.wrote || st.path != "out.txt" { t.Fatalf("expected write to out.txt") }
}

func TestTranscribeFile_PropagatesErrors(t *testing.T) {
    // decoder error
    uc1 := &TranscribeFile{Factory: func(string)(decoder.Decoder,error){ return nil, errors.New("boom") }}
    if _, err := uc1.Execute(context.Background(), TranscribeInput{Path:"x"}); err == nil { t.Fatal("expected error") }

    // repo error
    dec := &fakeDecoder{sr:8000, ch:1, pcm: make([]float32, 8000)}
    uc2 := &TranscribeFile{Factory: func(string)(decoder.Decoder,error){ return dec, nil }, Repo: &fakeRepo{err: errors.New("missing")}}
    if _, err := uc2.Execute(context.Background(), TranscribeInput{Path:"x"}); err == nil { t.Fatal("expected error") }

    // transcriber error
    uc3 := &TranscribeFile{Factory: func(string)(decoder.Decoder,error){ return dec, nil }, Repo: &fakeRepo{path:"/m"}, Trans: &fakeTranscriber{err: errors.New("oops")}}
    if _, err := uc3.Execute(context.Background(), TranscribeInput{Path:"x"}); err == nil { t.Fatal("expected error") }
}

// Integration tests with real MP3 decoder

func TestTranscribeFile_MP3_RealDecoder(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping MP3 integration test in short mode")
    }

    testMP3 := "../adapter/outbound/audio/decoder/testdata/test.mp3"
    // Check if test file exists
    if _, err := os.Stat(testMP3); os.IsNotExist(err) {
        t.Skip("test MP3 file not found")
    }

    // Use real decoder via Factory default (decoder.New)
    rep := &fakeRepo{path: "/models/ggml-tiny.en.bin"}
    tr := &fakeTranscriber{}
    st := &fakeStorage{}

    uc := &TranscribeFile{
        Repo:  rep,
        Trans: tr,
        Store: st,
        // Factory: nil -> uses real decoder.New()
    }

    result, err := uc.Execute(context.Background(), TranscribeInput{
        Path:      testMP3,
        ModelName: "ggml-tiny.en.bin",
        Language:  "en",
    })

    if err != nil {
        t.Fatalf("Execute failed for MP3: %v", err)
    }

    // Verify transcriber was called with valid PCM
    if tr.gotPCM == 0 {
        t.Error("Transcriber not called with PCM data")
    }

    // MP3 is ~3 seconds @ 48kHz, resampled to 16kHz = ~48000 samples
    if tr.gotPCM < 40000 || tr.gotPCM > 60000 {
        t.Errorf("unexpected PCM length: %d (expected ~48000)", tr.gotPCM)
    }

    t.Logf("MP3 decoded and resampled to %d samples @ 16kHz", tr.gotPCM)

    // Verify result
    if result.Language != "en" {
        t.Errorf("unexpected language: %s", result.Language)
    }
    if result.FullText != "hello world" {
        t.Errorf("unexpected text: %s", result.FullText)
    }
}

func TestTranscribeFile_MP3_InvalidFile(t *testing.T) {
    // Create invalid MP3 file
    invalidMP3 := t.TempDir() + "/invalid.mp3"
    if err := os.WriteFile(invalidMP3, []byte("not an mp3 file, just text"), 0644); err != nil {
        t.Fatal(err)
    }

    rep := &fakeRepo{path: "/models/ggml-tiny.en.bin"}
    tr := &fakeTranscriber{}
    st := &fakeStorage{}

    uc := &TranscribeFile{
        Repo:  rep,
        Trans: tr,
        Store: st,
        // Factory: nil -> uses real decoder.New()
    }

    _, err := uc.Execute(context.Background(), TranscribeInput{
        Path:      invalidMP3,
        ModelName: "ggml-tiny.en.bin",
    })

    if err == nil {
        t.Error("expected error for invalid MP3")
    }

    t.Logf("Got expected error: %v", err)
}
