package clock

import (
    "time"
    "gosper/internal/port"
)

type SystemClock struct{}

var _ port.Clock = (*SystemClock)(nil)

func (SystemClock) Now() time.Time { return time.Now() }

