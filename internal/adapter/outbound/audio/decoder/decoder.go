package decoder

// Package decoder provides a uniform interface for decoding audio files
// into normalized mono float32 PCM.

import (
    "fmt"
    "path/filepath"
)

type Info struct {
    SampleRate int
    Channels   int
    Frames     int64
    Path       string
}

type Decoder interface {
    Info() Info
    // DecodeAll returns normalized mono f32 PCM in range [-1,1].
    DecodeAll() ([]float32, error)
    Close() error
}

// New returns a Decoder for the given file path based on extension.
func New(path string) (Decoder, error) {
    ext := filepath.Ext(path)
    switch ext {
    case ".wav", ".Wave", ".WAV":
        return NewWAV(path)
	case ".mp3", ".MP3", ".Mp3":
        return NewMP3(path)
    default:
        return nil, fmt.Errorf("unsupported audio format: %s", ext)
    }
}

