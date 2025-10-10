package decoder

import "testing"

func TestDownmixToMonoF32_StereoAverage(t *testing.T) {
    // Interleaved stereo: L=1.0, R=0.0 repeated
    in := []float32{1, 0, 1, 0, 1, 0, 1, 0}
    mono := DownmixToMonoF32(in, 2)
    if len(mono) != len(in)/2 { t.Fatalf("expected %d, got %d", len(in)/2, len(mono)) }
    for i, v := range mono {
        if v != 0.5 {
            t.Fatalf("frame %d expected 0.5, got %f", i, v)
        }
    }
}

func TestDownmixToMonoF32_MonoCopy(t *testing.T) {
    in := []float32{0.1, -0.2, 0.3}
    mono := DownmixToMonoF32(in, 1)
    if len(mono) != len(in) { t.Fatalf("len mismatch") }
    for i := range in {
        if in[i] != mono[i] { t.Fatalf("copy mismatch at %d", i) }
    }
}

func TestPeakNormalizeF32(t *testing.T) {
    in := []float32{2, -2, 0}
    out := PeakNormalizeF32(in)
    if out[0] != 1 || out[1] != -1 { t.Fatalf("expected [1 -1 0], got %v", out) }

    // Already normalized
    in2 := []float32{0.5, -0.5}
    out2 := PeakNormalizeF32(in2)
    if out2[0] != 0.5 || out2[1] != -0.5 { t.Fatalf("should be unchanged: %v", out2) }
}

