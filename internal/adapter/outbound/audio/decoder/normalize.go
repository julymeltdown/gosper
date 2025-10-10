package decoder

// DownmixToMonoF32 averages interleaved channels to mono.
// If channels == 1 it returns a copy of the input.
func DownmixToMonoF32(in []float32, channels int) []float32 {
    if channels <= 1 {
        out := make([]float32, len(in))
        copy(out, in)
        return out
    }
    nFrames := len(in) / channels
    out := make([]float32, nFrames)
    for f := 0; f < nFrames; f++ {
        sum := float32(0)
        base := f * channels
        for c := 0; c < channels; c++ {
            sum += in[base+c]
        }
        out[f] = sum / float32(channels)
    }
    return out
}

// PeakNormalizeF32 scales the slice so that the peak absolute value is 1.0.
// If max <= 1, returns a copy of input unchanged.
func PeakNormalizeF32(in []float32) []float32 {
    max := float32(0)
    for _, v := range in {
        if v < 0 { v = -v }
        if v > max { max = v }
    }
    out := make([]float32, len(in))
    copy(out, in)
    if max <= 1 || max == 0 { // already normalized or silent
        return out
    }
    scale := 1 / max
    for i := range out {
        out[i] *= scale
    }
    return out
}

