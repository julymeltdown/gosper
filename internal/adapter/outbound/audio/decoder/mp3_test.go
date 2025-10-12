package decoder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMP3Decoder_ValidFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping MP3 test in short mode")
	}

	testFile := "testdata/test.mp3"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test MP3 file not found, run: cp ../../../../test.mp3 testdata/")
	}

	dec, err := NewMP3(testFile)
	if err != nil {
		t.Fatalf("NewMP3 failed: %v", err)
	}
	defer dec.Close()

	info := dec.Info()
	t.Logf("MP3 Info: SampleRate=%d Hz, Channels=%d, Frames=%d",
		info.SampleRate, info.Channels, info.Frames)

	// Validate info
	if info.SampleRate <= 0 {
		t.Errorf("invalid sample rate: %d", info.SampleRate)
	}
	if info.Channels != 2 {
		t.Errorf("expected 2 channels (go-mp3 always stereo), got %d", info.Channels)
	}
	if info.Path != testFile {
		t.Errorf("expected path %q, got %q", testFile, info.Path)
	}

	// Decode audio
	samples, err := dec.DecodeAll()
	if err != nil {
		t.Fatalf("DecodeAll failed: %v", err)
	}

	if len(samples) == 0 {
		t.Fatal("no samples decoded")
	}

	t.Logf("Decoded %d mono samples", len(samples))

	// Verify normalization: all samples should be in [-1, 1]
	outOfRange := 0
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			outOfRange++
			if outOfRange <= 5 {
				t.Errorf("sample[%d] out of range: %f", i, s)
			}
		}
	}
	if outOfRange > 5 {
		t.Errorf("...and %d more samples out of range", outOfRange-5)
	}

	// Verify frame count was updated (for VBR MP3)
	info2 := dec.Info()
	if info2.Frames <= 0 {
		t.Errorf("frame count not updated after decode: %d", info2.Frames)
	}
}

func TestMP3Decoder_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.mp3")
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewMP3(tmpFile)
	if err == nil {
		t.Error("expected error for empty file")
	}
	if err != nil && !contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error, got: %v", err)
	}
}

func TestMP3Decoder_InvalidFormat(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.mp3")
	if err := os.WriteFile(tmpFile, []byte("not an mp3 file, just text"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewMP3(tmpFile)
	if err == nil {
		t.Error("expected error for invalid MP3 format")
	}
	if err != nil && !contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' in error, got: %v", err)
	}
}

func TestMP3Decoder_OversizedFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large file test in short mode")
	}

	tmpFile := filepath.Join(t.TempDir(), "large.mp3")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	// Create 201MB file (exceeds 200MB limit)
	if err := f.Truncate(201 * 1024 * 1024); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	_, err = NewMP3(tmpFile)
	if err == nil {
		t.Error("expected error for oversized file")
	}
	if err != nil && !contains(err.Error(), "too large") {
		t.Errorf("expected 'too large' in error, got: %v", err)
	}
}

func TestMP3Decoder_NonExistentFile(t *testing.T) {
	_, err := NewMP3("/nonexistent/path/file.mp3")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if err != nil && !contains(err.Error(), "open") {
		t.Errorf("expected 'open' in error, got: %v", err)
	}
}

func TestMP3Decoder_Close(t *testing.T) {
	testFile := "testdata/test.mp3"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test MP3 file not found")
	}

	dec, err := NewMP3(testFile)
	if err != nil {
		t.Fatalf("NewMP3 failed: %v", err)
	}

	// Close once
	if err := dec.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Close again (should not panic, may return error)
	if err := dec.Close(); err != nil {
		t.Logf("Second Close returned error (acceptable): %v", err)
	}
}

func TestNew_MP3(t *testing.T) {
	testFile := "testdata/test.mp3"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test MP3 file not found")
	}

	// Test factory function
	dec, err := New(testFile)
	if err != nil {
		t.Fatalf("New failed for .mp3 file: %v", err)
	}
	defer dec.Close()

	info := dec.Info()
	if info.SampleRate <= 0 {
		t.Error("invalid sample rate from factory")
	}

	// Verify it's an MP3 decoder
	if _, ok := dec.(*mp3Decoder); !ok {
		t.Errorf("expected *mp3Decoder, got %T", dec)
	}
}

func TestNew_MP3_CaseInsensitive(t *testing.T) {
	// Test that factory recognizes .mp3, .MP3, .Mp3 extensions
	testCases := []struct {
		ext  string
		name string
	}{
		{".mp3", "lowercase"},
		{".MP3", "uppercase"},
		{".Mp3", "mixed case"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test"+tc.ext)
			// Write minimal MP3 header (will fail decode, but that's ok)
			// We're just testing extension matching
			if err := os.WriteFile(tmpFile, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := New(tmpFile)
			// Should attempt MP3 decoder (will fail on invalid data)
			if err == nil {
				t.Error("expected error for dummy file")
			}
			// Error should mention "mp3" or "invalid"
			if err != nil && !contains(err.Error(), "mp3") && !contains(err.Error(), "invalid") {
				t.Errorf("unexpected error for %s: %v", tc.ext, err)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
