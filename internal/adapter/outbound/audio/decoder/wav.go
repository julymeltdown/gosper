package decoder

import (
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "math"
    "os"
)

// Minimal WAV decoder supporting PCM16 and PCM32 float, mono or stereo.

type wavDecoder struct {
    f    *os.File
    info Info
    dataOffset int64
    dataSize   int64
    audioFormat uint16
    _bitsPerSample uint16
}

func NewWAV(path string) (Decoder, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    wd := &wavDecoder{f: f}
    if err := wd.readHeader(path); err != nil {
        f.Close()
        return nil, err
    }
    return wd, nil
}

func (w *wavDecoder) Info() Info { return w.info }

func (w *wavDecoder) Close() error { return w.f.Close() }

func (w *wavDecoder) DecodeAll() ([]float32, error) {
    if _, err := w.f.Seek(w.dataOffset, io.SeekStart); err != nil { return nil, err }
    r := io.LimitReader(w.f, w.dataSize)
    frames := int(w.dataSize) / (w.info.Channels * (w.bitsPerSample()/8))
    // read interleaved, downmix, normalize to f32
    switch w.audioFormat {
    case 1: // PCM
        if w.bitsPerSample() != 16 { return nil, fmt.Errorf("unsupported PCM bits per sample: %d", w.bitsPerSample()) }
        buf := make([]byte, w.info.Channels*2)
        out := make([]float32, 0, frames)
        for {
            if _, err := io.ReadFull(r, buf); err != nil {
                if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) { break }
                return nil, err
            }
            // downmix
            sum := int32(0)
            for c := 0; c < w.info.Channels; c++ {
                off := c*2
                v := int16(binary.LittleEndian.Uint16(buf[off:]))
                sum += int32(v)
            }
            avg := float32(sum) / float32(w.info.Channels)
            out = append(out, avg/32768.0)
        }
        return out, nil
    case 3: // IEEE float32
        if w.bitsPerSample() != 32 { return nil, fmt.Errorf("unsupported float bits per sample: %d", w.bitsPerSample()) }
        buf := make([]byte, w.info.Channels*4)
        tmp := make([]float32, w.info.Channels)
        out := make([]float32, 0, frames)
        for {
            if _, err := io.ReadFull(r, buf); err != nil {
                if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) { break }
                return nil, err
            }
            for c := 0; c < w.info.Channels; c++ {
                bits := binary.LittleEndian.Uint32(buf[c*4:])
                tmp[c] = math.Float32frombits(bits)
            }
            // downmix average
            sum := float32(0)
            for c := 0; c < w.info.Channels; c++ { sum += tmp[c] }
            out = append(out, sum/float32(w.info.Channels))
        }
        return PeakNormalizeF32(out), nil
    default:
        return nil, fmt.Errorf("unsupported audio format: %d", w.audioFormat)
    }
}

// internal WAV parsing state
func (w *wavDecoder) bitsPerSample() int { return int(w._bitsPerSample) }
func (w *wavDecoder) readHeader(path string) error {
    var riff [4]byte
    if _, err := io.ReadFull(w.f, riff[:]); err != nil { return err }
    if string(riff[:]) != "RIFF" { return fmt.Errorf("not a RIFF file") }
    if _, err := w.f.Seek(4, io.SeekCurrent); err != nil { return err } // skip chunk size
    var wave [4]byte
    if _, err := io.ReadFull(w.f, wave[:]); err != nil { return err }
    if string(wave[:]) != "WAVE" { return fmt.Errorf("not a WAVE file") }

    // read chunks until we find 'fmt ' and 'data'
    var haveFmt, haveData bool
    for {
        var id [4]byte
        if _, err := io.ReadFull(w.f, id[:]); err != nil { return err }
        var size uint32
        if err := binary.Read(w.f, binary.LittleEndian, &size); err != nil { return err }
        switch string(id[:]) {
        case "fmt ":
            if size < 16 { return fmt.Errorf("fmt chunk too small") }
            var audioFormat uint16
            var numChannels uint16
            var sampleRate uint32
            var byteRate uint32
            var blockAlign uint16
            var bitsPerSample uint16
            if err := binary.Read(w.f, binary.LittleEndian, &audioFormat); err != nil { return err }
            if err := binary.Read(w.f, binary.LittleEndian, &numChannels); err != nil { return err }
            if err := binary.Read(w.f, binary.LittleEndian, &sampleRate); err != nil { return err }
            if err := binary.Read(w.f, binary.LittleEndian, &byteRate); err != nil { return err }
            if err := binary.Read(w.f, binary.LittleEndian, &blockAlign); err != nil { return err }
            if err := binary.Read(w.f, binary.LittleEndian, &bitsPerSample); err != nil { return err }
            // skip any extra bytes
            if rem := int64(size) - 16; rem > 0 {
                if _, err := w.f.Seek(rem, io.SeekCurrent); err != nil { return err }
            }
            w.audioFormat = audioFormat
            w._bitsPerSample = bitsPerSample
            w.info.SampleRate = int(sampleRate)
            w.info.Channels = int(numChannels)
            haveFmt = true
        case "data":
            w.dataOffset, _ = w.f.Seek(0, io.SeekCurrent)
            w.dataSize = int64(size)
            if _, err := w.f.Seek(int64(size), io.SeekCurrent); err != nil { return err }
            haveData = true
        default:
            // skip chunk
            if _, err := w.f.Seek(int64(size), io.SeekCurrent); err != nil { return err }
        }
        if haveFmt && haveData {
            // compute frames
            bps := int(w._bitsPerSample) / 8
            if bps == 0 || w.info.Channels == 0 { return fmt.Errorf("invalid wav header") }
            w.info.Frames = int64(w.dataSize) / int64(w.info.Channels*bps)
            w.info.Path = path
            return nil
        }
    }
}

// fields needed but not exposed
// no-op, placeholder for future extensions
