package errors

import "errors"

var (
    InvalidArgs        = errors.New("invalid arguments")
    AudioError         = errors.New("audio error")
    ModelError         = errors.New("model error")
    TranscriptionError = errors.New("transcription error")
    FsError            = errors.New("filesystem error")
)

// Wrap labels an error with a sentinel cause for errors.Is checks.
func Wrap(cause, err error) error {
    if err == nil { return nil }
    return struct{ error }{error: join(cause, err)}
}

func join(a, b error) error { return errors.Join(a, b) }

