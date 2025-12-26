package config

import (
    "encoding/json"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
)

// Config holds runtime configuration resolved from flags/env.
type Config struct {
    Model       string
    Language    string
    Threads     int
    CacheDir    string
    LogLevel    string
    LastDeviceID string
    AudioFeedback bool
    OutputDeviceID string
    BeepVolume     float64
}

func FromEnv() Config {
    c := Config{
        Model:    getenv("GOSPER_MODEL", "ggml-base.en.bin"),
        Language: getenv("GOSPER_LANG", "auto"),
        CacheDir: getenv("GOSPER_CACHE", ""),
        LogLevel: getenv("GOSPER_LOG", "info"),
        AudioFeedback: getenv("GOSPER_AUDIO_FEEDBACK", "") == "1",
        OutputDeviceID: getenv("GOSPER_OUTPUT_DEVICE", ""),
    }
    if v := getenv("GOSPER_THREADS", ""); v != "" {
        if n, err := strconv.Atoi(v); err == nil { c.Threads = n }
    }
    if v := getenv("GOSPER_BEEP_VOLUME", ""); v != "" {
        if n, err := strconv.ParseFloat(v, 64); err == nil { c.BeepVolume = n }
    }
    return c
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
    if p := os.Getenv("GOSPER_CONFIG"); p != "" { return p }
    if d, err := os.UserConfigDir(); err == nil {
        return filepath.Join(d, "gosper", "config.json")
    }
    if runtime.GOOS == "windows" {
        if d := os.Getenv("APPDATA"); d != "" {
            return filepath.Join(d, "gosper", "config.json")
        }
    }
    return filepath.Join(os.TempDir(), "gosper", "config.json")
}

func LoadFile(path string) (Config, error) {
    var c Config = FromEnv()
    b, err := os.ReadFile(path)
    if err != nil { return c, err }
    var f Config
    if err := json.Unmarshal(b, &f); err != nil { return c, err }
    // merge env defaults with file values (file takes precedence for stored fields)
    if f.Model != "" { c.Model = f.Model }
    if f.Language != "" { c.Language = f.Language }
    if f.Threads != 0 { c.Threads = f.Threads }
    if f.CacheDir != "" { c.CacheDir = f.CacheDir }
    if f.LogLevel != "" { c.LogLevel = f.LogLevel }
    if f.LastDeviceID != "" { c.LastDeviceID = f.LastDeviceID }
    c.AudioFeedback = f.AudioFeedback
    if f.OutputDeviceID != "" { c.OutputDeviceID = f.OutputDeviceID }
    if f.BeepVolume != 0 { c.BeepVolume = f.BeepVolume }
    return c, nil
}

func SaveFile(path string, c Config) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { return err }
    b, err := json.MarshalIndent(c, "", "  ")
    if err != nil { return err }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, b, 0o644); err != nil { return err }
    return os.Rename(tmp, path)
}
