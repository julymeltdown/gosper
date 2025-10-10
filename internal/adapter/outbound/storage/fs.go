package storage

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"

    "gosper/internal/domain"
    "gosper/internal/port"
)

type FS struct{}

var _ port.Storage = (*FS)(nil)

func (FS) WriteFile(ctx context.Context, path string, r io.Reader) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
    tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
    if err != nil { return err }
    defer func(){ tmp.Close(); os.Remove(tmp.Name()) }()
    if _, err := io.Copy(tmp, r); err != nil { return err }
    if err := tmp.Sync(); err != nil { return err }
    if err := tmp.Close(); err != nil { return err }
    return os.Rename(tmp.Name(), path)
}

func (FS) WriteTranscript(ctx context.Context, path string, t domain.Transcript) error {
    ext := filepath.Ext(path)
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
    switch ext {
    case ".json":
        b, err := json.MarshalIndent(t, "", "  ")
        if err != nil { return err }
        return writeBytesAtomic(path, b)
    case ".txt", "":
        s := t.FullText
        return writeBytesAtomic(path, []byte(s))
    default:
        return fmt.Errorf("unsupported transcript format: %s", ext)
    }
}

func (FS) TempPath(ctx context.Context, pattern string) (string, error) {
    f, err := os.CreateTemp(os.TempDir(), pattern)
    if err != nil { return "", err }
    path := f.Name()
    f.Close()
    return path, nil
}

func writeBytesAtomic(path string, b []byte) error {
    tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
    if err != nil { return err }
    _, werr := tmp.Write(b)
    serr := tmp.Sync()
    cerr := tmp.Close()
    if werr != nil { os.Remove(tmp.Name()); return werr }
    if serr != nil { os.Remove(tmp.Name()); return serr }
    if cerr != nil { os.Remove(tmp.Name()); return cerr }
    return os.Rename(tmp.Name(), path)
}

