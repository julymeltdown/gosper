package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gosper/internal/usecase"
)

// Logger interface for dependency injection
type Logger interface {
	Println(v ...interface{})
}

// Server handles HTTP requests
type Server struct {
	transcribeUC *usecase.TranscribeFile
	logger       Logger
	httpServer   *http.Server
}

// Config holds server configuration
type Config struct {
	Addr          string
	ModelDefault  string
	LanguageDefault string
}

// NewServer creates a new HTTP server
func NewServer(transcribeUC *usecase.TranscribeFile, logger Logger, cfg Config) *Server {
	s := &Server{
		transcribeUC: transcribeUC,
		logger:       logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/api/transcribe", s.transcribeHandler(cfg))

	s.httpServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: corsMiddleware(mux),
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Println("HTTP server listening on", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Println("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// transcribeHandler handles transcription requests
func (s *Server) transcribeHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.clientError(w, r, http.StatusMethodNotAllowed, "POST required")
			return
		}

		tmp, cleanup, err := s.handleFileUpload(w, r)
		if err != nil {
			return
		}
		defer cleanup()

		modelName := r.FormValue("model")
		if modelName == "" {
			modelName = cfg.ModelDefault
		}
		lang := r.FormValue("lang")
		if lang == "" {
			lang = cfg.LanguageDefault
		}

		start := time.Now()
		tr, trErr := s.transcribeUC.Execute(r.Context(), usecase.TranscribeInput{
			Path:      tmp.Name(),
			ModelName: modelName,
			Language:  lang,
		})
		dur := time.Since(start)
		if trErr != nil {
			s.serverError(w, r, trErr)
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
}

// handleFileUpload processes multipart file upload
func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request) (*os.File, func(), error) {
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB
		s.clientError(w, r, http.StatusBadRequest, fmt.Sprintf("parse form: %v", err))
		return nil, nil, err
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		s.clientError(w, r, http.StatusBadRequest, "missing file field 'audio'")
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
		s.serverError(w, r, fmt.Errorf("tmp: %v", err))
		return nil, nil, err
	}
	cleanup := func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}

	if _, err := io.Copy(tmp, file); err != nil {
		cleanup()
		s.serverError(w, r, fmt.Errorf("write: %v", err))
		return nil, nil, err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		cleanup()
		s.serverError(w, r, fmt.Errorf("seek: %v", err))
		return nil, nil, err
	}

	// Convert WebM to WAV if needed
	if ext == ".webm" || ext == ".ogg" {
		wavTmp, convertErr := s.convertToWAV(tmp.Name())
		if convertErr != nil {
			cleanup()
			s.serverError(w, r, fmt.Errorf("convert to wav: %v", convertErr))
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
func (s *Server) convertToWAV(inputPath string) (*os.File, error) {
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

// Error handling

type responseError struct {
	Error string `json:"error"`
}

func (s *Server) errorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	env := responseError{Error: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Println(r.Method, r.URL.Path, r.RemoteAddr, "error:", err)
	message := "the server encountered a problem and could not process your request"
	s.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (s *Server) clientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	s.errorResponse(w, r, status, message)
}

// Default logger implementation
type DefaultLogger struct {
	*log.Logger
}

func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		Logger: log.New(os.Stdout, "[HTTP] ", log.LstdFlags),
	}
}
