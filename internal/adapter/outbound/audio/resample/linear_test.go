package resample

import (
    "math"
    "testing"
)

func genSine(sr, hz int, durSec float64, amp float32) []float32 {
    n := int(float64(sr) * durSec)
    out := make([]float32, n)
    for i := 0; i < n; i++ {
        out[i] = amp * float32(math.Sin(2*math.Pi*float64(hz)*float64(i)/float64(sr)))
    }
    return out
}

func maxAbs(x []float32) float32 {
    m := float32(0)
    for _, v := range x {
        if v < 0 { v = -v }
        if v > m { m = v }
    }
    return m
}

func TestLinear_Identity(t *testing.T) {
    in := genSine(16000, 440, 0.1, 0.8)
    out := Linear(in, 16000, 16000)
    if len(out) != len(in) { t.Fatalf("len mismatch") }
    if maxAbs(out)-maxAbs(in) > 1e-6 { t.Fatalf("amplitude changed") }
}

func TestLinear_UpsampleAndDownsample(t *testing.T) {
    in := genSine(16000, 440, 0.2, 0.7)
    up := Linear(in, 16000, 48000)
    if len(up) <= len(in) { t.Fatalf("expected upsampled length > in") }
    down := Linear(up, 48000, 16000)
    if len(down) != len(in) && (len(down) < len(in)-2 || len(down) > len(in)+2) {
        t.Fatalf("unexpected down len: %d vs %d", len(down), len(in))
    }
    // Amplitude should be close
    diff := math.Abs(float64(maxAbs(down)-maxAbs(in)))
    if diff > 0.05 { // allow some tolerance
        t.Fatalf("amplitude drift too high: %f", diff)
    }
}

func TestLinear_ConstantSignal(t *testing.T) {
    in := make([]float32, 1000)
    for i := range in { in[i] = 0.5 }
    out := Linear(in, 22050, 16000)
    if len(out) == 0 { t.Fatalf("no output") }
    // Check first and last sample approx 0.5
    if math.Abs(float64(out[0]-0.5)) > 1e-3 || math.Abs(float64(out[len(out)-1]-0.5)) > 1e-3 {
        t.Fatalf("constant signal changed: first=%f last=%f", out[0], out[len(out)-1])
    }
}

func TestLinear_ShortInputs(t *testing.T) {
    // len=1
    in := []float32{0.7}
    out := Linear(in, 44100, 16000)
    if len(out) == 0 { t.Fatalf("expected at least 1 sample") }
    // len=2
    in2 := []float32{0.0, 1.0}
    out2 := Linear(in2, 8000, 16000)
    if out2[0] != 0.0 { t.Fatalf("first sample mismatch") }
}
