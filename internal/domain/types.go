package domain

// TranscriptSegment represents a segment of transcribed audio.
type TranscriptSegment struct {
    Index   int
    StartMS int64
    EndMS   int64
    Text    string
}

// Transcript is the full transcription result.
type Transcript struct {
    Language string
    Segments []TranscriptSegment
    FullText string
}

// ModelConfig holds model/runtime parameters for transcription.
type ModelConfig struct {
    ModelName     string
    ModelPath     string
    Language      string
    Translate     bool
    Threads       uint
    Timestamps    bool
    BeamSize      int
    MaxTokens     uint
    InitialPrompt string
}

// AudioFormat describes captured/decoded audio format.
type AudioFormat struct {
    SampleRate int    // Hz
    Channels   int    // 1 or 2
    SampleType string // "f32" or "i16"
}

// Device represents an audio capture device.
type Device struct {
    ID   string
    Name string
}

