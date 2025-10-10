package model

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type FSRepo struct {
    BaseURL   string            // optional remote base
    CacheDir  string            // defaults to user cache dir
    Checksums map[string]string // optional sha256 for model files
    Retry     int               // download retries (default 2)
}

func (r *FSRepo) Ensure(ctx context.Context, modelName string) (string, error) {
    if modelName == "" { modelName = "ggml-tiny.en.bin" }
    // If modelName is a path that exists, return it
    if fi, err := os.Stat(modelName); err == nil && fi.Mode().IsRegular() {
        abs, _ := filepath.Abs(modelName)
        return abs, nil
    }
    // Resolve cache path
    cacheDir := r.CacheDir
    if cacheDir == "" {
        if d, err := os.UserCacheDir(); err == nil {
            cacheDir = filepath.Join(d, "gosper", "models")
        } else {
            cacheDir = filepath.Join(os.TempDir(), "gosper", "models")
        }
    }
    if err := os.MkdirAll(cacheDir, 0o755); err != nil {
        return "", err
    }
    local := filepath.Join(cacheDir, modelName)
    if fi, err := os.Stat(local); err == nil && fi.Mode().IsRegular() {
        if ok, _ := r.verify(local, modelName); ok {
            return local, nil
        }
        // bad checksum: remove and redownload
        _ = os.Remove(local)
    }
    // Attempt download if BaseURL configured
    if r.BaseURL == "" {
        return "", fmt.Errorf("model %q not found locally; set BaseURL to enable download", modelName)
    }
    url := fmt.Sprintf("%s/%s", trimSlash(r.BaseURL), modelName)
    // Download with simple client
    attempts := r.Retry
    if attempts <= 0 { attempts = 2 }
    var lastErr error
    for i := 0; i < attempts; i++ {
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil { return "", err }
        resp, err := http.DefaultClient.Do(req)
        if err != nil { lastErr = err; continue }
        func() {
            defer resp.Body.Close()
            if resp.StatusCode != http.StatusOK { lastErr = fmt.Errorf("download failed: %s", resp.Status); return }
            tmp := local + ".part"
            if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil { lastErr = err; return }
            f, err := os.Create(tmp)
            if err != nil { lastErr = err; return }
            defer func(){ f.Close(); os.Remove(tmp) }()
            if _, err := ioCopy(ctx, f, resp.Body); err != nil { lastErr = err; return }
            if err := f.Sync(); err != nil { lastErr = err; return }
            if err := f.Close(); err != nil { lastErr = err; return }
            if err := os.Rename(tmp, local); err != nil { lastErr = err; return }
            lastErr = nil
        }()
        if lastErr == nil {
            // verify checksum if available
            if ok, verr := r.verify(local, modelName); verr == nil && ok {
                return local, nil
            } else if r.Checksums != nil { // checksum configured but mismatch
                _ = os.Remove(local)
                lastErr = fmt.Errorf("checksum mismatch for %s", modelName)
            } else {
                return local, nil
            }
        }
        // simple backoff
        select { case <-time.After(time.Duration(i+1)*time.Second): case <-ctx.Done(): return "", ctx.Err() }
    }
    return "", lastErr
}

func trimSlash(s string) string {
    for len(s) > 0 && s[len(s)-1] == '/' { s = s[:len(s)-1] }
    return s
}

// Simplified io copy with context; avoids pulling extra deps.
func ioCopy(ctx context.Context, dst *os.File, src io.Reader) (int64, error) {
    buf := make([]byte, 1<<20)
    var n int64
    for {
        select { case <-ctx.Done(): return n, ctx.Err(); default: }
        nr, er := src.Read(buf)
        if nr > 0 {
            nw, ew := dst.Write(buf[:nr])
            if nw > 0 { n += int64(nw) }
            if ew != nil { return n, ew }
            if nr != nw { return n, io.ErrShortWrite }
        }
        if er != nil { if er == io.EOF { break } ; return n, er }
    }
    return n, nil
}

// verify checks sha256 if provided in FSRepo.Checksums; if no checksum available, returns (true, nil)
func (r *FSRepo) verify(path, name string) (bool, error) {
    sum, ok := r.Checksums[name]
    if !ok || sum == "" { return true, nil }
    got, err := fileSha256(path)
    if err != nil { return false, err }
    return strings.EqualFold(got, sum), nil
}

func fileSha256(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil { return "", err }
    defer f.Close()
    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil { return "", err }
    return hex.EncodeToString(h.Sum(nil)), nil
}
