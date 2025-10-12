# Configuration Guide

This guide covers all configuration options for Gosper, including environment variables, config files, model management, and advanced tuning.

## Table of Contents

- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Model Management](#model-management)
- [Audio Format Specifications](#audio-format-specifications)
- [Performance Tuning](#performance-tuning)
- [Server Configuration](#server-configuration)

## Environment Variables

Gosper can be configured via environment variables. These take precedence over config file settings.

### Core Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `GOSPER_MODEL` | string | `ggml-tiny.en.bin` | Model name or absolute path |
| `GOSPER_LANG` | string | `auto` | Language code (en, es, fr, etc.) or `auto` |
| `GOSPER_THREADS` | int | CPU cores | Number of threads for inference |
| `GOSPER_CACHE` | string | OS cache dir | Model cache directory |
| `GOSPER_LOG` | string | `info` | Log level: `debug`, `info`, `warn`, `error` |

### Audio Settings (CLI)

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `GOSPER_AUDIO_FEEDBACK` | bool | `false` | Enable beep on recording start/stop |
| `GOSPER_OUTPUT_DEVICE` | string | `default` | Audio output device ID for beeps |
| `GOSPER_BEEP_VOLUME` | float | `0.5` | Beep volume (0.0 - 1.0) |

### Server Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PORT` | int | `8080` | HTTP server port |
| `HOST` | string | `0.0.0.0` | HTTP server bind address |
| `MODEL_BASE_URL` | string | Hugging Face | Base URL for model downloads |

### Examples

**Basic CLI Usage**:
```bash
export GOSPER_MODEL=ggml-base.en.bin
export GOSPER_LANG=en
export GOSPER_THREADS=4

gosper transcribe meeting.wav
```

**Server Configuration**:
```bash
export PORT=9000
export GOSPER_MODEL=ggml-medium.en.bin
export GOSPER_CACHE=/var/cache/gosper/models
export GOSPER_LOG=debug

./server
```

**Development Setup**:
```bash
export GOSPER_LOG=debug
export GOSPER_MODEL_PATH=/path/to/models/ggml-tiny.en.bin
export GOSPER_INTEGRATION=1  # Enable integration tests

make test
make itest
```

## Configuration File

Gosper reads configuration from `~/.config/gosper/config.json` (or `$XDG_CONFIG_HOME/gosper/config.json`).

### File Location

**Linux/macOS**:
```
~/.config/gosper/config.json
```

**Windows**:
```
%APPDATA%\gosper\config.json
```

### Format

```json
{
  "model": "ggml-base.en.bin",
  "lang": "en",
  "threads": 4,
  "cache_dir": "/path/to/models",
  "log_level": "info",
  "LastDeviceID": "default",
  "AudioFeedback": true,
  "OutputDeviceID": "speakers",
  "BeepVolume": 0.5
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `model` | string | Default model name or path |
| `lang` | string | Default language code |
| `threads` | int | Thread count for inference |
| `cache_dir` | string | Model cache directory |
| `log_level` | string | Logging level |
| `LastDeviceID` | string | Last used audio input device |
| `AudioFeedback` | bool | Enable recording beeps |
| `OutputDeviceID` | string | Audio output device for beeps |
| `BeepVolume` | float | Beep volume (0.0 - 1.0) |

### Precedence Order

Configuration is loaded in this order (later overrides earlier):

1. **Default values** (hardcoded in binary)
2. **Config file** (`~/.config/gosper/config.json`)
3. **Environment variables** (`GOSPER_*`)
4. **Command-line flags** (`--model`, `--lang`, etc.)

**Example**:
```bash
# Config file says: model=ggml-tiny.en.bin
# Environment says: GOSPER_MODEL=ggml-base.en.bin
# Command-line says: --model ggml-medium.en.bin

# Result: ggml-medium.en.bin (command-line wins)
```

## Model Management

### Model Selection

**Option 1: Model Name** (auto-download from Hugging Face):
```bash
gosper transcribe audio.mp3 --model ggml-base.en.bin
# Downloads to cache if not found
```

**Option 2: Absolute Path**:
```bash
gosper transcribe audio.mp3 --model /path/to/ggml-large-v3.bin
# Uses local file directly
```

**Option 3: Environment Variable**:
```bash
export GOSPER_MODEL=ggml-medium.en.bin
gosper transcribe audio.mp3
```

### Model Cache Directory

**Default Locations**:
- **Linux**: `~/.cache/gosper/`
- **macOS**: `~/Library/Caches/gosper/`
- **Windows**: `%LOCALAPPDATA%\gosper\cache\`

**Custom Cache**:
```bash
export GOSPER_CACHE=/var/cache/gosper
gosper transcribe audio.mp3
```

### Model Download Behavior

1. **Check if model path is absolute** → use directly
2. **Check cache directory** → use if found
3. **Download from MODEL_BASE_URL** → save to cache
4. **Verify SHA256 checksum** (optional)
5. **Retry with exponential backoff** on failure

### Model Sources

**Default Source** (Hugging Face):
```
https://huggingface.co/ggerganov/whisper.cpp/resolve/main/
```

**Custom Source**:
```bash
export MODEL_BASE_URL=https://your-cdn.com/models/
gosper transcribe audio.mp3 --model ggml-base.en.bin
# Downloads from: https://your-cdn.com/models/ggml-base.en.bin
```

### Available Models

| Model | Size | English-Only | Multilingual | RAM Required | Typical Speed |
|-------|------|--------------|--------------|--------------|---------------|
| `ggml-tiny.en.bin` | 75 MB | ✅ | ❌ | 500 MB | 5x real-time |
| `ggml-tiny.bin` | 75 MB | ❌ | ✅ | 500 MB | 5x real-time |
| `ggml-base.en.bin` | 142 MB | ✅ | ❌ | 800 MB | 3x real-time |
| `ggml-base.bin` | 142 MB | ❌ | ✅ | 800 MB | 3x real-time |
| `ggml-small.en.bin` | 466 MB | ✅ | ❌ | 1.5 GB | 1.5x real-time |
| `ggml-small.bin` | 466 MB | ❌ | ✅ | 1.5 GB | 1.5x real-time |
| `ggml-medium.en.bin` | 1.5 GB | ✅ | ❌ | 3 GB | 0.5x real-time |
| `ggml-medium.bin` | 1.5 GB | ❌ | ✅ | 3 GB | 0.5x real-time |
| `ggml-large-v3.bin` | 3.1 GB | ❌ | ✅ | 6 GB | 0.25x real-time |

**Notes**:
- `.en.bin` models are English-only (faster, more accurate for English)
- Non-`.en` models support 100+ languages
- Speed estimates are approximate (varies by hardware)
- RAM includes model + decoded audio buffer

### Manual Model Download

```bash
# Create cache directory
mkdir -p ~/.cache/gosper

# Download tiny model (fast, English only)
curl -L -o ~/.cache/gosper/ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Download base model (balanced, English only)
curl -L -o ~/.cache/gosper/ggml-base.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin

# Download large model (best accuracy, multilingual)
curl -L -o ~/.cache/gosper/ggml-large-v3.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin
```

### SHA256 Verification

Gosper optionally verifies model checksums during download.

**Expected Checksums**:
```bash
# ggml-tiny.en.bin
sha256sum ~/.cache/gosper/ggml-tiny.en.bin
# Expected: (check whisper.cpp repository for latest)

# Verify manually
curl -L https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin.sha256
```

## Audio Format Specifications

### WAV Format

**Supported**:
- **Extensions**: `.wav`, `.Wave`, `.WAV` (case-insensitive)
- **Sample Rates**: 8000 - 96000 Hz
- **Channels**: Mono (1) or Stereo (2)
- **Bit Depth**: 16-bit PCM or 32-bit IEEE float
- **File Size**: No limit

**Processing**:
1. Decode PCM samples to float32
2. Downmix stereo to mono (if stereo): `(L + R) / 2`
3. Resample to 16000 Hz (Whisper requirement)

**Example**:
```bash
# 44.1kHz stereo WAV → 16kHz mono for Whisper
gosper transcribe music.wav --lang en
```

### MP3 Format

**Supported**:
- **Extensions**: `.mp3`, `.MP3` (case-insensitive)
- **Sample Rates**: 8000 - 96000 Hz
- **Channels**: Mono or Stereo
- **Bitrate**: All bitrates (CBR, VBR, ABR)
- **File Size**: **Maximum 200 MB** compressed

**Limitations**:
- **200 MB limit** to prevent memory exhaustion
- Decoded audio requires ~3x compressed size in memory
- VBR MP3s may not report accurate duration until fully decoded

**Processing**:
1. Validate file size (reject if > 200 MB)
2. Decode MP3 to 16-bit stereo PCM
3. Convert to float32: `value / 32768.0`
4. Downmix stereo to mono
5. Resample to 16000 Hz

**For Large Files**:
```bash
# If MP3 > 200MB, convert to WAV first
ffmpeg -i large-podcast.mp3 large-podcast.wav
gosper transcribe large-podcast.wav
```

### Format Detection

Format is detected by **file extension** (case-insensitive):

```go
switch ext {
case ".wav", ".Wave", ".WAV":
    return NewWAV(path)
case ".mp3", ".MP3":
    return NewMP3(path)
default:
    return nil, fmt.Errorf("unsupported format: %s", ext)
}
```

**Unsupported Formats**:
Convert using ffmpeg:
```bash
# M4A → MP3
ffmpeg -i audio.m4a audio.mp3

# FLAC → WAV
ffmpeg -i audio.flac audio.wav

# OGG → MP3
ffmpeg -i audio.ogg audio.mp3
```

## Performance Tuning

### Thread Count

**Automatic** (default):
```bash
# Uses all available CPU cores
gosper transcribe audio.mp3
```

**Manual**:
```bash
# Use 4 threads
gosper transcribe audio.mp3 --threads 4

# Use 8 threads
export GOSPER_THREADS=8
gosper transcribe audio.mp3
```

**Guidelines**:
- **2-4 threads**: Typical for small models (tiny, base)
- **4-8 threads**: Optimal for medium models
- **8-16 threads**: Large models on high-end CPUs
- **More threads ≠ always faster** (diminishing returns after 8)

### Model Selection Strategy

| Use Case | Model | Reason |
|----------|-------|--------|
| **Quick testing** | `ggml-tiny.en.bin` | Fastest, good enough for demos |
| **Production (English)** | `ggml-base.en.bin` | Balanced speed/accuracy |
| **High accuracy (English)** | `ggml-medium.en.bin` | Best for English |
| **Multilingual** | `ggml-small.bin` | Good balance for 100+ languages |
| **Maximum accuracy** | `ggml-large-v3.bin` | Slowest but most accurate |

### Memory Optimization

**Estimated Memory Usage**:
```
Total RAM = Model Size + Audio Buffer + Overhead

Examples:
- tiny.en + 10 min audio ≈ 75 MB + 200 MB + 50 MB = 325 MB
- medium.en + 1 hour audio ≈ 1.5 GB + 1.2 GB + 300 MB ≈ 3 GB
```

**For Large Audio Files**:
```bash
# Process in chunks (future feature)
# Currently: use smaller model or add more RAM
```

### Language Detection

**Automatic** (default):
```bash
gosper transcribe audio.mp3 --lang auto
# Whisper detects language (adds ~1s overhead)
```

**Explicit** (faster):
```bash
gosper transcribe audio.mp3 --lang en
# Skips detection, 10-20% faster
```

**Supported Languages** (multilingual models only):
English, Spanish, French, German, Italian, Portuguese, Dutch, Russian, Chinese, Japanese, Korean, Arabic, and 80+ more.

## Server Configuration

### Basic Setup

```bash
# Default (port 8080, all interfaces)
./server

# Custom port
export PORT=9000
./server

# Bind to localhost only
export HOST=127.0.0.1
./server
```

### Docker Configuration

```bash
docker run -p 8080:8080 \
  -e GOSPER_MODEL=ggml-base.en.bin \
  -e GOSPER_THREADS=4 \
  -e GOSPER_LOG=info \
  -v gosper-models:/root/.cache/gosper \
  gosper/server:latest
```

### Kubernetes Configuration

**ConfigMap** (`config.yaml`):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gosper-config
data:
  GOSPER_MODEL: "ggml-base.en.bin"
  GOSPER_THREADS: "4"
  GOSPER_LOG: "info"
  PORT: "8080"
```

**Deployment** (reference ConfigMap):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gosper-be
spec:
  template:
    spec:
      containers:
      - name: server
        image: gosper/server:local
        envFrom:
        - configMapRef:
            name: gosper-config
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
```

### Resource Limits

**Kubernetes Resource Requests** (recommended):

| Model | Memory Request | Memory Limit | CPU Request | CPU Limit |
|-------|----------------|--------------|-------------|-----------|
| `tiny` | 1 GB | 2 GB | 500m | 1000m |
| `base` | 2 GB | 4 GB | 1000m | 2000m |
| `small` | 4 GB | 8 GB | 2000m | 4000m |
| `medium` | 8 GB | 16 GB | 2000m | 4000m |
| `large` | 16 GB | 32 GB | 4000m | 8000m |

**Docker Memory Limits**:
```bash
docker run -p 8080:8080 \
  --memory=4g \
  --cpus=2 \
  -e GOSPER_MODEL=ggml-base.en.bin \
  gosper/server:latest
```

## Advanced Topics

### Custom Model Base URL

Host models on your own CDN or file server:

```bash
export MODEL_BASE_URL=https://models.yourcompany.com/whisper/

# Downloads from: https://models.yourcompany.com/whisper/ggml-base.en.bin
gosper transcribe audio.mp3 --model ggml-base.en.bin
```

### Logging Configuration

**Log Levels**:
```bash
export GOSPER_LOG=debug   # Verbose debugging
export GOSPER_LOG=info    # Normal operation (default)
export GOSPER_LOG=warn    # Warnings only
export GOSPER_LOG=error   # Errors only
```

**JSON Logging** (for structured logging):
```bash
# Future feature - currently plain text
export GOSPER_LOG_FORMAT=json
```

### Audio Device Management (CLI)

**List Devices**:
```bash
gosper devices list
```

**Select Device**:
```bash
# By ID
gosper record --device "hw:0,0"

# By name (fuzzy match)
gosper record --device "USB Microphone"
```

**Device Selection Algorithm**:
1. Exact ID match
2. Exact name match (case-insensitive)
3. Prefix match
4. Substring match
5. Fuzzy match (Levenshtein distance)

**Persist Selection**:
Last used device is saved to `~/.config/gosper/config.json`

## Next Steps

- **[API Reference](API.md)** - HTTP API endpoints and examples
- **[Quick Start](QUICKSTART.md)** - Get started quickly
- **[Deployment](DEPLOYMENT.md)** - Production k8s deployment
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions
