package decoder

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/go-mp3"
)

type mp3Decoder struct {
	f       *os.File
	decoder *mp3.Decoder
	info    Info
}

// NewMP3 creates a new MP3 decoder for the given file path.
// Returns an error if the file doesn't exist, is empty, too large (>200MB),
// or contains invalid MP3 data.
//
// The decoder always outputs mono float32 PCM samples normalized to [-1, 1].
// Sample rate is preserved from the source MP3 file.
func NewMP3(path string) (Decoder, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("mp3: open: %w", err)
	}

	// Validate file size before decoding
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("mp3: stat: %w", err)
	}

	if stat.Size() == 0 {
		f.Close()
		return nil, fmt.Errorf("mp3: empty file")
	}

	// Prevent abuse: reject files > 200MB compressed
	// (decoded size will be ~3x larger)
	const maxCompressedSize = 200 * 1024 * 1024
	if stat.Size() > maxCompressedSize {
		f.Close()
		return nil, fmt.Errorf("mp3: file too large (%d MB, max 200 MB)",
			stat.Size()/1024/1024)
	}

	// Create MP3 decoder
	dec, err := mp3.NewDecoder(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("mp3: invalid format: %w", err)
	}

	// Validate sample rate
	sr := dec.SampleRate()
	if sr <= 0 || sr > 96000 {
		f.Close()
		return nil, fmt.Errorf("mp3: invalid sample rate %d Hz (expected 8000-96000)", sr)
	}

	md := &mp3Decoder{
		f:       f,
		decoder: dec,
	}

	md.info.Path = path
	md.info.SampleRate = sr
	md.info.Channels = 2 // go-mp3 always outputs stereo

	// Try to get frame count (may be -1 for VBR MP3)
	length := dec.Length()
	if length > 0 {
		// Length is in bytes, each sample is 4 bytes (16-bit stereo)
		md.info.Frames = length / 4
	} else {
		// VBR MP3: frame count unknown until decode
		md.info.Frames = -1
	}

	return md, nil
}

func (m *mp3Decoder) Info() Info {
	return m.info
}

func (m *mp3Decoder) Close() error {
	// Note: go-mp3 decoder doesn't have Close method
	return m.f.Close()
}

// DecodeAll reads and decodes the entire MP3 file,
// converting it to mono float32 PCM samples normalized to [-1, 1].
//
// The go-mp3 library always outputs 16-bit little-endian stereo,
// which we then downmix to mono using the existing utility function.
func (m *mp3Decoder) DecodeAll() ([]float32, error) {
	// Read all decoded data
	// go-mp3 outputs 16-bit LE stereo (4 bytes per sample)
	data, err := io.ReadAll(m.decoder)
	if err != nil {
		return nil, fmt.Errorf("mp3: read: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("mp3: no audio data")
	}

	// Validate data size (must be multiple of 4 bytes)
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("mp3: incomplete sample data (size=%d bytes)", len(data))
	}

	numSamples := len(data) / 4
	stereo := make([]float32, numSamples*2) // interleaved stereo

	// Convert int16 stereo to float32 stereo
	for i := 0; i < numSamples; i++ {
		left := int16(binary.LittleEndian.Uint16(data[i*4 : i*4+2]))
		right := int16(binary.LittleEndian.Uint16(data[i*4+2 : i*4+4]))

		// Normalize int16 to float32 [-1, 1]
		stereo[i*2] = float32(left) / 32768.0
		stereo[i*2+1] = float32(right) / 32768.0
	}

	// Use existing utility for downmix (consistent with codebase)
	mono := DownmixToMonoF32(stereo, 2)

	// Update frame count if it was unknown (VBR MP3)
	if m.info.Frames < 0 {
		m.info.Frames = int64(len(mono))
	}

	return mono, nil
}

