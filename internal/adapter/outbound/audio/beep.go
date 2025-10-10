package audio

import (
    "fmt"
    "math"
    "time"
)

// ConsoleBell emits a simple terminal bell character as a lightweight feedback.
func ConsoleBell() { fmt.Print("\a") }

// Tone generates a sine wave at given frequency and duration, mono f32 [-1,1].
// Useful if later routed to an audio output adapter.
func Tone(frequency float64, duration time.Duration, sampleRate int) []float32 {
    n := int(float64(sampleRate) * duration.Seconds())
    out := make([]float32, n)
    ang := 2 * math.Pi * frequency / float64(sampleRate)
    for i := 0; i < n; i++ { out[i] = 0.2 * float32(math.Sin(ang*float64(i))) }
    return out
}

