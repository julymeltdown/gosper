package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gosper/internal/adapter/outbound/model"
	"gosper/internal/adapter/outbound/storage"
	"gosper/internal/adapter/outbound/whispercpp"
	"gosper/internal/config"
	"gosper/internal/usecase"
)

type application struct {
	cfg      *config.Config
	logger   *log.Logger
	usecases struct {
		transcribeFile *usecase.TranscribeFile
	}
}

// Server is the main application server.
type Server struct {
	server *http.Server
	app    *application
}

// NewServer creates a new server.
func NewServer(app *application) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/transcribe", app.transcribeHandler)

	return &Server{
		server: &http.Server{
			Addr:    app.cfg.Addr,
			Handler: corsMiddleware(mux),
		},
		app: app,
	}
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() error {
	log.Printf("server listening on %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("shutting down server")
	return s.server.Shutdown(ctx)
}

func main() {
	cfg := config.FromEnv()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	app := &application{
		cfg:    cfg,
		logger: logger,
		usecases: struct{ transcribeFile *usecase.TranscribeFile }{
			transcribeFile: &usecase.TranscribeFile{
				Repo:  &model.FSRepo{BaseURL: cfg.ModelBaseURL},
				Trans: &whispercpp.Transcriber{},
				Store: storage.FS{},
			},
		},
	}

	srv := NewServer(app)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow requests from any origin (adjust for production)
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

type responseError struct{ Error string `json:"error"` }

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := responseError{Error: message.(string)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Println(r.Method, r.URL.Path, r.RemoteAddr, "error:", err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) clientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	app.errorResponse(w, r, status, message)
}

func (app *application) transcribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed, "POST required")
		return
	}

	tmp, cleanup, err := app.handleFileUpload(w, r)
	if err != nil {
		return
	}
	defer cleanup()

	modelName := r.FormValue("model")
	if modelName == "" {
		modelName = app.cfg.Model
	}
	lang := r.FormValue("lang")
	if lang == "" {
		lang = app.cfg.Language
	}

	start := time.Now()
	tr, trErr := app.usecases.transcribeFile.Execute(r.Context(), usecase.TranscribeInput{
		Path:      tmp.Name(),
		ModelName: modelName,
		Language:  lang,
	})
	dur := time.Since(start)
	if trErr != nil {
		app.serverError(w, r, trErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"language":    tr.Language,
		"text":        tr.FullText,
		"segments":    tr.Segments,
		"duration_ms": dur.Milliseconds(),
	})
}

func (app *application) handleFileUpload(w http.ResponseWriter, r *http.Request) (*os.File, func(), error) {
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB
		app.clientError(w, r, http.StatusBadRequest, fmt.Sprintf("parse form: %v", err))
		return nil, nil, err
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest, "missing file field 'audio'")
		return nil, nil, err
	}
	defer file.Close()

	// Persist to temp file for decoder compatibility
	tmpDir := os.TempDir()
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".wav"
	}

	// Normalize extension
	ext = strings.ToLower(ext)

	tmp, err := os.CreateTemp(tmpDir, "upload-*"+ext)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("tmp: %v", err))
		return nil, nil, err
	}
	cleanup := func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}

	if _, err := io.Copy(tmp, file); err != nil {
		cleanup()
		app.serverError(w, r, fmt.Errorf("write: %v", err))
		return nil, nil, err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		cleanup()
		app.serverError(w, r, fmt.Errorf("seek: %v", err))
		return nil, nil, err
	}

	// Convert WebM to WAV if needed
	if ext == ".webm" || ext == ".ogg" {
		wavTmp, convertErr := app.convertToWAV(tmp.Name())
		if convertErr != nil {
			cleanup()
			app.serverError(w, r, fmt.Errorf("convert to wav: %v", convertErr))
			return nil, nil, convertErr
		}

		// Clean up original file, return converted WAV
		cleanup()
		newCleanup := func() {
			wavTmp.Close()
			os.Remove(wavTmp.Name())
		}
		return wavTmp, newCleanup, nil
	}

	return tmp, cleanup, nil
}

// convertToWAV converts an audio file to WAV format using ffmpeg
func (app *application) convertToWAV(inputPath string) (*os.File, error) {
	tmpDir := os.TempDir()
	wavFile, err := os.CreateTemp(tmpDir, "converted-*.wav")
	if err != nil {
		return nil, fmt.Errorf("create temp wav: %v", err)
	}
	wavFile.Close() // ffmpeg will write to it

	// Convert to 16kHz mono WAV (whisper requirement)
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		"-y", // overwrite
		wavFile.Name(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(wavFile.Name())
		return nil, fmt.Errorf("ffmpeg: %v, output: %s", err, string(output))
	}

	// Reopen for reading
	f, err := os.Open(wavFile.Name())
	if err != nil {
		os.Remove(wavFile.Name())
		return nil, fmt.Errorf("reopen wav: %v", err)
	}

	return f, nil
}

