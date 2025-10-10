package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestFSRepo_Ensure_LocalPath tests that Ensure returns the path directly
// when given an existing absolute path
func TestFSRepo_Ensure_LocalPath(t *testing.T) {
	// Create a temp model file
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "ggml-tiny.en.bin")
	if err := os.WriteFile(modelPath, []byte("fake model"), 0644); err != nil {
		t.Fatal(err)
	}

	repo := &FSRepo{}
	got, err := repo.Ensure(context.Background(), modelPath)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	// Should return absolute path
	if !filepath.IsAbs(got) {
		t.Errorf("Expected absolute path, got %q", got)
	}

	// Should point to the same file
	if got != modelPath {
		absPath, _ := filepath.Abs(modelPath)
		if got != absPath {
			t.Errorf("Ensure() = %q, want %q", got, absPath)
		}
	}
}

// TestFSRepo_Ensure_CachedModel tests that Ensure returns cached path
// when model already exists in cache
func TestFSRepo_Ensure_CachedModel(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatal(err)
	}

	modelName := "ggml-tiny.en.bin"
	cachedPath := filepath.Join(cacheDir, modelName)

	// Pre-populate cache
	if err := os.WriteFile(cachedPath, []byte("cached model"), 0644); err != nil {
		t.Fatal(err)
	}

	repo := &FSRepo{
		CacheDir: cacheDir,
	}

	got, err := repo.Ensure(context.Background(), modelName)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if got != cachedPath {
		t.Errorf("Ensure() = %q, want %q", got, cachedPath)
	}

	// Verify file content
	content, _ := os.ReadFile(got)
	if string(content) != "cached model" {
		t.Errorf("Expected cached content")
	}
}

// TestFSRepo_Ensure_DownloadSuccess tests successful model download
func TestFSRepo_Ensure_DownloadSuccess(t *testing.T) {
	// Create mock HTTP server
	modelContent := []byte("downloaded model content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ggml-tiny.en.bin" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(modelContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	repo := &FSRepo{
		BaseURL:  server.URL,
		CacheDir: cacheDir,
		Retry:    1,
	}

	modelName := "ggml-tiny.en.bin"
	got, err := repo.Ensure(context.Background(), modelName)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	// Should return cached path
	expectedPath := filepath.Join(cacheDir, modelName)
	if got != expectedPath {
		t.Errorf("Ensure() = %q, want %q", got, expectedPath)
	}

	// Verify downloaded content
	content, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != string(modelContent) {
		t.Errorf("Downloaded content mismatch")
	}
}

// TestFSRepo_Ensure_ChecksumValidation tests SHA256 verification
func TestFSRepo_Ensure_ChecksumValidation(t *testing.T) {
	modelContent := []byte("test model content")
	hash := sha256.Sum256(modelContent)
	correctChecksum := hex.EncodeToString(hash[:])

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(modelContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	t.Run("valid checksum", func(t *testing.T) {
		repo := &FSRepo{
			BaseURL:  server.URL,
			CacheDir: cacheDir,
			Checksums: map[string]string{
				"model.bin": correctChecksum,
			},
			Retry: 1,
		}

		_, err := repo.Ensure(context.Background(), "model.bin")
		if err != nil {
			t.Errorf("Expected success with valid checksum, got error: %v", err)
		}
	})

	t.Run("invalid checksum", func(t *testing.T) {
		// Clean cache
		os.RemoveAll(cacheDir)

		repo := &FSRepo{
			BaseURL:  server.URL,
			CacheDir: cacheDir,
			Checksums: map[string]string{
				"model2.bin": "invalid_checksum_1234567890abcdef",
			},
			Retry: 1,
		}

		_, err := repo.Ensure(context.Background(), "model2.bin")
		if err == nil {
			t.Error("Expected error with invalid checksum")
		}
	})
}

// TestFSRepo_Ensure_RetryOnFailure tests retry logic
func TestFSRepo_Ensure_RetryOnFailure(t *testing.T) {
	attempts := 0
	modelContent := []byte("model after retry")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// Fail first request
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on retry
		w.WriteHeader(http.StatusOK)
		w.Write(modelContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	repo := &FSRepo{
		BaseURL:  server.URL,
		CacheDir: cacheDir,
		Retry:    2, // Allow 2 retries
	}

	got, err := repo.Ensure(context.Background(), "model.bin")
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts (1 fail + 1 success), got %d", attempts)
	}

	// Verify content
	content, _ := os.ReadFile(got)
	if string(content) != string(modelContent) {
		t.Error("Content mismatch after retry")
	}
}

// TestFSRepo_Ensure_EmptyModelName tests default model handling
func TestFSRepo_Ensure_EmptyModelName(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")

	// Create default model in cache
	defaultModel := "ggml-tiny.en.bin"
	cachedPath := filepath.Join(cacheDir, defaultModel)
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(cachedPath, []byte("default"), 0644)

	repo := &FSRepo{
		CacheDir: cacheDir,
	}

	// Empty model name should use default
	got, err := repo.Ensure(context.Background(), "")
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if got != cachedPath {
		t.Errorf("Ensure() = %q, want default %q", got, cachedPath)
	}
}

// TestFileSha256 tests SHA256 calculation
func TestFileSha256(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content for hashing")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	hash := sha256.Sum256(content)
	expectedHash := hex.EncodeToString(hash[:])

	got, err := fileSha256(testFile)
	if err != nil {
		t.Fatalf("fileSha256() error = %v", err)
	}

	if got != expectedHash {
		t.Errorf("fileSha256() = %q, want %q", got, expectedHash)
	}
}

// Note: verify() is tested indirectly through TestFSRepo_Ensure_ChecksumValidation

// TestTrimSlash tests URL path normalization
func TestTrimSlash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com", "https://example.com"},
		{"/path/to/file/", "/path/to/file"},
		{"no-slash", "no-slash"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimSlash(tt.input)
			if got != tt.want {
				t.Errorf("trimSlash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
