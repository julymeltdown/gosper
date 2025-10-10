package resample

// Linear resampler from inRate to outRate for mono float32 PCM.
// Not optimized; sufficient for correctness and tests.
func Linear(pcm []float32, inRate, outRate int) []float32 {
    if inRate == outRate || len(pcm) == 0 {
        out := make([]float32, len(pcm))
        copy(out, pcm)
        return out
    }
    ratio := float64(outRate) / float64(inRate)
    outLen := int(float64(len(pcm)) * ratio)
    if outLen <= 1 {
        outLen = 1
    }
    out := make([]float32, outLen)
    for i := 0; i < outLen; i++ {
        // compute position in input
        pos := float64(i) / ratio
        ip := int(pos)
        frac := float32(pos - float64(ip))
        if ip >= len(pcm)-1 {
            out[i] = pcm[len(pcm)-1]
            continue
        }
        a := pcm[ip]
        b := pcm[ip+1]
        out[i] = a + (b-a)*frac
    }
    return out
}

