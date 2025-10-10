package audio

import "gosper/internal/port"

// NewInput returns the platform audio input implementation if available.
// Default build (no 'malgo' tag) uses stub.
func NewInput() port.AudioInput { return stubInput{} }

