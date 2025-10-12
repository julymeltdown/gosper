//go:build !malgo

package audio
import "time"

// Default PlayBeep uses console bell.
func PlayBeep() { ConsoleBell() }

type BeepOptions struct {
    DeviceID string
    Volume   float32
    Freq     float64
    Duration time.Duration
}

// PlayBeepOptions (console build) ignores options and rings bell.
func PlayBeepOptions(BeepOptions) { ConsoleBell() }
