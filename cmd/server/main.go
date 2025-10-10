package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "gosper/internal/adapter/outbound/model"
    "gosper/internal/adapter/outbound/storage"
    "gosper/internal/adapter/outbound/whispercpp"
    "gosper/internal/usecase"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })
    mux.HandleFunc("/api/transcribe", transcribeHandler)

    addr := ":8080"
    if p := os.Getenv("PORT"); p != "" { addr = ":" + p }
    log.Printf("server listening on %s", addr)
    srv := &http.Server{Addr: addr, Handler: mux}
    log.Fatal(srv.ListenAndServe())
}

type responseError struct{ Error string `json:"error"` }

func jsonError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(responseError{Error: msg})
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        jsonError(w, http.StatusMethodNotAllowed, "POST required")
        return
    }
    if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB
        jsonError(w, http.StatusBadRequest, fmt.Sprintf("parse form: %v", err))
        return
    }
    file, header, err := r.FormFile("audio")
    if err != nil { jsonError(w, http.StatusBadRequest, "missing file field 'audio'"); return }
    defer file.Close()

    // Persist to temp file for decoder compatibility
    tmpDir := os.TempDir()
    ext := filepath.Ext(header.Filename)
    if ext == "" { ext = ".wav" }
    tmp, err := os.CreateTemp(tmpDir, "upload-*"+ext)
    if err != nil { jsonError(w, http.StatusInternalServerError, fmt.Sprintf("tmp: %v", err)); return }
    defer func() { tmp.Close(); os.Remove(tmp.Name()) }()
    if _, err := io.Copy(tmp, file); err != nil { jsonError(w, http.StatusInternalServerError, fmt.Sprintf("write: %v", err)); return }
    if _, err := tmp.Seek(0, 0); err != nil { jsonError(w, http.StatusInternalServerError, fmt.Sprintf("seek: %v", err)); return }

    // Build use case
    uc := &usecase.TranscribeFile{
        Repo:  &model.FSRepo{BaseURL: os.Getenv("MODEL_BASE_URL")},
        Trans: &whispercpp.Transcriber{},
        Store: storage.FS{},
    }

    modelName := r.FormValue("model")
    if modelName == "" { modelName = os.Getenv("GOSPER_MODEL") }
    lang := r.FormValue("lang")
    if lang == "" { lang = os.Getenv("GOSPER_LANG") }

    start := time.Now()
    tr, trErr := uc.Execute(r.Context(), usecase.TranscribeInput{
        Path:      tmp.Name(),
        ModelName: modelName,
        Language:  lang,
    })
    dur := time.Since(start)
    if trErr != nil {
        jsonError(w, http.StatusBadGateway, trErr.Error())
        return
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "language":   tr.Language,
        "text":       tr.FullText,
        "segments":   tr.Segments,
        "duration_ms": dur.Milliseconds(),
    })
}

