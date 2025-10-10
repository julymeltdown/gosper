//go:build malgo

package audio

import "gosper/internal/port"

func NewInput() port.AudioInput { return MalgoInput{} }

