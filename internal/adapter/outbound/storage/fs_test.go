package storage

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "strings"

    
    "gosper/internal/domain"
)

func TestStorageFS_JSONAndTXT(t *testing.T) {
    dir := t.TempDir()
    st := FS{}
    tr := domain.Transcript{Language: "en", FullText: "hello", Segments: []domain.TranscriptSegment{{Index:0, Text:"hello"}}}

    jsonPath := filepath.Join(dir, "out.json")
    if err := st.WriteTranscript(context.Background(), jsonPath, tr); err != nil { t.Fatalf("json write: %v", err) }
    b, _ := os.ReadFile(jsonPath)
    if !contains(string(b), "\"FullText\"") { t.Fatalf("json missing FullText: %s", string(b)) }

    txtPath := filepath.Join(dir, "out.txt")
    if err := st.WriteTranscript(context.Background(), txtPath, tr); err != nil { t.Fatalf("txt write: %v", err) }
    tb, _ := os.ReadFile(txtPath)
    if string(tb) != "hello" { t.Fatalf("unexpected txt: %q", string(tb)) }
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }

// TestStorageFS_WriteFile tests writing arbitrary files
func TestStorageFS_WriteFile(t *testing.T) {
	dir := t.TempDir()
	st := FS{}

	t.Run("write new file", func(t *testing.T) {
		path := filepath.Join(dir, "test.dat")
		content := strings.NewReader("test content data")

		err := st.WriteFile(context.Background(), path, content)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Verify file was written
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(data) != "test content data" {
			t.Errorf("Content = %q, want %q", string(data), "test content data")
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		path := filepath.Join(dir, "existing.dat")
		os.WriteFile(path, []byte("old content"), 0644)

		content := strings.NewReader("new content")
		err := st.WriteFile(context.Background(), path, content)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != "new content" {
			t.Errorf("Content = %q, want %q", string(data), "new content")
		}
	})

	t.Run("create subdirectories", func(t *testing.T) {
		path := filepath.Join(dir, "sub", "dir", "file.dat")
		content := strings.NewReader("nested")

		err := st.WriteFile(context.Background(), path, content)
		if err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("File was not created in nested directory")
		}
	})
}

// TestStorageFS_TempPath tests temporary file creation
func TestStorageFS_TempPath(t *testing.T) {
	st := FS{}

	t.Run("create temp file", func(t *testing.T) {
		path, err := st.TempPath(context.Background(), "test-*")
		if err != nil {
			t.Fatalf("TempPath() error = %v", err)
		}

		// Verify path is not empty
		if path == "" {
			t.Error("TempPath() returned empty path")
		}

		// Verify file exists
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Temp file does not exist: %v", err)
		}

		// Clean up
		os.Remove(path)
	})

	t.Run("pattern is used", func(t *testing.T) {
		path, err := st.TempPath(context.Background(), "myprefix-*.tmp")
		if err != nil {
			t.Fatalf("TempPath() error = %v", err)
		}

		filename := filepath.Base(path)
		if !strings.HasPrefix(filename, "myprefix-") {
			t.Errorf("Filename %q doesn't match pattern", filename)
		}
		if !strings.HasSuffix(filename, ".tmp") {
			t.Errorf("Filename %q doesn't match pattern", filename)
		}

		os.Remove(path)
	})

	t.Run("multiple calls return different paths", func(t *testing.T) {
		path1, _ := st.TempPath(context.Background(), "test-*")
		path2, _ := st.TempPath(context.Background(), "test-*")

		if path1 == path2 {
			t.Error("TempPath() returned same path twice")
		}

		os.Remove(path1)
		os.Remove(path2)
	})
}

// TestStorageFS_WriteTranscript_EdgeCases tests edge cases
func TestStorageFS_WriteTranscript_EdgeCases(t *testing.T) {
	dir := t.TempDir()
	st := FS{}

	t.Run("empty transcript", func(t *testing.T) {
		path := filepath.Join(dir, "empty.txt")
		tr := domain.Transcript{Language: "", FullText: ""}

		err := st.WriteTranscript(context.Background(), path, tr)
		if err != nil {
			t.Fatalf("WriteTranscript() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != "" {
			t.Errorf("Expected empty content")
		}
	})

	t.Run("transcript with special characters", func(t *testing.T) {
		path := filepath.Join(dir, "special.txt")
		tr := domain.Transcript{
			Language: "en",
			FullText: "Hello \"world\"! 日本語 🎉",
		}

		err := st.WriteTranscript(context.Background(), path, tr)
		if err != nil {
			t.Fatalf("WriteTranscript() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != tr.FullText {
			t.Errorf("Content = %q, want %q", string(data), tr.FullText)
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		path := filepath.Join(dir, "unsupported.xml")
		tr := domain.Transcript{FullText: "test"}

		err := st.WriteTranscript(context.Background(), path, tr)
		if err == nil {
			t.Error("Expected error for unsupported format")
		}
		if !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("Error message should mention unsupported format, got: %v", err)
		}
	})

	t.Run("long transcript", func(t *testing.T) {
		path := filepath.Join(dir, "long.txt")
		longText := strings.Repeat("Lorem ipsum dolor sit amet. ", 10000)
		tr := domain.Transcript{FullText: longText}

		err := st.WriteTranscript(context.Background(), path, tr)
		if err != nil {
			t.Fatalf("WriteTranscript() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != longText {
			t.Error("Long text content mismatch")
		}
	})
}
