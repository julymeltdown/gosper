package logadapter

import (
    "context"
    "log/slog"

    "gosper/internal/port"
)

type SlogLogger struct{
    l *slog.Logger
}

func NewSlogLogger(l *slog.Logger) *SlogLogger { return &SlogLogger{l: l} }

var _ port.Logger = (*SlogLogger)(nil)

func (s *SlogLogger) Debug(ctx context.Context, msg string, kv ...any) { s.l.DebugContext(ctx, msg, kv...) }
func (s *SlogLogger) Info(ctx context.Context, msg string, kv ...any)  { s.l.InfoContext(ctx, msg, kv...) }
func (s *SlogLogger) Warn(ctx context.Context, msg string, kv ...any)  { s.l.WarnContext(ctx, msg, kv...) }
func (s *SlogLogger) Error(ctx context.Context, msg string, kv ...any) { s.l.ErrorContext(ctx, msg, kv...) }

