# Quick Start Guide

Get started with Gosper in minutes. This guide covers the fastest way to start transcribing audio files.

## Table of Contents

- [Docker Quick Start](#docker-quick-start)
- [CLI Quick Start](#cli-quick-start)
- [HTTP API Quick Start](#http-api-quick-start)
- [Web UI Quick Start](#web-ui-quick-start)
- [First Transcription](#your-first-transcription)
- [Next Steps](#next-steps)

## Docker Quick Start

The fastest way to try Gosper:

```bash
# Run the server
docker run -p 8080:8080 gosper/server:latest

# In another terminal, transcribe an audio file
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@your-audio.mp3" \
  -F "lang=auto"
```

**That's it!** Your transcript will be returned as JSON.

### With Custom Model

```bash
# Use a larger model for better accuracy
docker run -p 8080:8080 \
  -e GOSPER_MODEL=ggml-base.en.bin \
  gosper/server:latest

# Transcribe
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@meeting.wav" \
  -F "model=ggml-base.en.bin" \
  -F "lang=en"
```

### Persist Models

Cache models between container restarts:

```bash
# Create volume for model cache
docker volume create gosper-models

# Run with volume
docker run -p 8080:8080 \
  -v gosper-models:/root/.cache/gosper \
  gosper/server:latest
```

## CLI Quick Start

### Install

**From Release**:
```bash
# Linux/macOS
curl -L -o gosper https://github.com/yourusername/gosper/releases/latest/download/gosper-$(uname -s)-$(uname -m)
chmod +x gosper
sudo mv gosper /usr/local/bin/

# Verify
gosper --version
```

**From Source**:
```bash
git clone https://github.com/yourusername/gosper.git
cd gosper
make deps
make build
sudo cp dist/gosper /usr/local/bin/
```

### Basic Usage

```bash
# Transcribe a WAV file
gosper transcribe meeting.wav --lang en

# Transcribe an MP3 file with automatic language detection
gosper transcribe podcast.mp3 --lang auto

# Use a specific model
gosper transcribe interview.wav \
  --model ggml-base.en.bin \
  --lang en

# Save to file
gosper transcribe lecture.mp3 \
  --lang en \
  -o transcript.txt
```

### Record from Microphone

```bash
# Record 30 seconds and transcribe
gosper record --duration 30s

# Record with audio feedback (beep)
gosper record --duration 30s --audio-feedback

# Save transcript to file
gosper record --duration 1m -o notes.txt
```

## HTTP API Quick Start

### Start Server

**Using Binary**:
```bash
# Build server
go build -tags "whisper" -o server ./cmd/server

# Run
./server
# Listening on :8080
```

**Using Docker**:
```bash
docker run -p 8080:8080 gosper/server:latest
```

### API Examples

**Basic Transcription**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@audio.mp3" \
  -F "lang=auto"
```

**With Specific Model**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@audio.wav" \
  -F "model=ggml-tiny.en.bin" \
  -F "lang=en"
```

**Response Format**:
```json
{
  "text": "This is the transcribed text from your audio file.",
  "language": "en",
  "duration_ms": 1250,
  "segments": [
    {
      "start_ms": 0,
      "end_ms": 1250,
      "text": "This is the transcribed text from your audio file."
    }
  ]
}
```

## Web UI Quick Start

### Access UI

1. Start the server (Docker or binary)
2. Open browser to `http://localhost:8080/`
3. Upload an audio file (WAV or MP3)
4. Click "Transcribe"
5. View results

### Kubernetes/k3s Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment with k3s.

**Quick k3s Deploy**:
```bash
# Build images
docker build -f Dockerfile.server -t gosper/server:local .
docker build -f Dockerfile.frontend -t gosper/fe:local .

# Import to k3s
docker save gosper/server:local | sudo k3s ctr images import -
docker save gosper/fe:local | sudo k3s ctr images import -

# Deploy
bash scripts/k3s/deploy.sh
```

Access:
- Backend API: `http://<NODE_IP>:31209/api/transcribe`
- Frontend UI: `http://<NODE_IP>:31987/`

## Your First Transcription

### Step 1: Get Sample Audio

```bash
# Download JFK sample from whisper.cpp
curl -L -o jfk.wav \
  https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav
```

### Step 2: Transcribe

**Using CLI**:
```bash
gosper transcribe jfk.wav --lang en
```

**Using API**:
```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@jfk.wav" \
  -F "lang=en"
```

### Step 3: View Results

**CLI Output**:
```
Transcribing jfk.wav...
Language: en
Duration: 11.0s

Transcript:
And so my fellow Americans ask not what your country can do for you
ask what you can do for your country.

Saved to: jfk.txt
```

**API Response**:
```json
{
  "text": "And so my fellow Americans ask not what your country can do for you ask what you can do for your country.",
  "language": "en",
  "duration_ms": 11000,
  "segments": [
    {
      "start_ms": 0,
      "end_ms": 5500,
      "text": "And so my fellow Americans ask not what your country can do for you"
    },
    {
      "start_ms": 5500,
      "end_ms": 11000,
      "text": "ask what you can do for your country."
    }
  ]
}
```

## Common Use Cases

### Transcribe Multiple Files

```bash
# Bash loop
for file in recordings/*.mp3; do
  gosper transcribe "$file" -o "transcripts/$(basename "$file" .mp3).txt"
done
```

### Batch API Requests

```bash
# Parallel transcription with GNU parallel
ls recordings/*.wav | parallel -j 4 \
  'curl -X POST http://localhost:8080/api/transcribe -F "audio=@{}" -F "lang=auto" > {.}.json'
```

### Real-Time Transcription

```bash
# Record and transcribe continuously (30s chunks)
while true; do
  gosper record --duration 30s -o "notes-$(date +%s).txt"
done
```

### Language Detection

```bash
# Auto-detect language
gosper transcribe multilingual-audio.mp3 --lang auto

# Force specific language
gosper transcribe spanish-audio.mp3 --lang es
```

## Configuration

### Environment Variables

```bash
# Set default model
export GOSPER_MODEL=ggml-base.en.bin

# Set cache directory
export GOSPER_CACHE=/path/to/models

# Set thread count
export GOSPER_THREADS=4

# Set log level
export GOSPER_LOG=debug

# Transcribe
gosper transcribe audio.mp3
```

### Config File

Create `~/.config/gosper/config.json`:

```json
{
  "model": "ggml-base.en.bin",
  "lang": "en",
  "threads": 4,
  "cache_dir": "/path/to/models"
}
```

## Supported Audio Formats

### WAV
- **Extensions**: `.wav`, `.Wave`, `.WAV`
- **Sample Rates**: 8000-96000 Hz
- **Channels**: Mono or stereo
- **Bit Depth**: 16-bit PCM or 32-bit float
- **File Size**: No limit

### MP3
- **Extensions**: `.mp3`, `.MP3`
- **Sample Rates**: 8000-96000 Hz
- **Channels**: Mono or stereo
- **Bitrate**: All bitrates supported (CBR, VBR, ABR)
- **File Size**: Maximum 200 MB

Both formats are automatically:
- Resampled to 16 kHz (Whisper requirement)
- Downmixed to mono (if stereo)
- Normalized to float32

## Available Models

| Model | Size | Accuracy | Speed | Use Case |
|-------|------|----------|-------|----------|
| `ggml-tiny.en.bin` | 75 MB | Low | Very Fast | Quick testing, English only |
| `ggml-base.en.bin` | 142 MB | Medium | Fast | General purpose, English only |
| `ggml-small.en.bin` | 466 MB | Good | Medium | High accuracy, English only |
| `ggml-medium.en.bin` | 1.5 GB | High | Slow | Production use, English only |
| `ggml-large-v3.bin` | 3.1 GB | Highest | Very Slow | Maximum accuracy, multilingual |

**Download Models**:
```bash
# Tiny (fast, English only)
curl -L -o ggml-tiny.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.en.bin

# Base (balanced, English only)
curl -L -o ggml-base.en.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin
```

Models are automatically downloaded on first use if not found locally.

## Performance Tips

### Speed Optimization

```bash
# Use more threads (default: number of CPU cores)
gosper transcribe audio.mp3 --threads 8

# Use smaller model
gosper transcribe audio.mp3 --model ggml-tiny.en.bin
```

### Accuracy Optimization

```bash
# Use larger model
gosper transcribe audio.mp3 --model ggml-medium.en.bin

# Specify language (skip detection)
gosper transcribe audio.mp3 --lang en
```

### Memory Management

```bash
# For large MP3 files (>200MB), convert to WAV first
ffmpeg -i large-audio.mp3 large-audio.wav
gosper transcribe large-audio.wav
```

## Troubleshooting

### Model Not Found

**Error**: `model not found: ggml-tiny.en.bin`

**Solution**: Download manually or let Gosper download automatically:
```bash
gosper transcribe audio.mp3 --model ggml-tiny.en.bin
# Model will be downloaded to cache directory
```

### Audio Format Not Supported

**Error**: `unsupported audio format: .m4a`

**Solution**: Convert to WAV or MP3:
```bash
ffmpeg -i audio.m4a audio.mp3
gosper transcribe audio.mp3
```

### File Too Large

**Error**: `mp3: file too large (250 MB, max 200 MB)`

**Solution**: Convert to WAV (no size limit):
```bash
ffmpeg -i large-file.mp3 large-file.wav
gosper transcribe large-file.wav
```

### Microphone Not Found

**Error**: `audio device not found`

**Solution** (macOS): Grant microphone permission
- System Settings → Privacy & Security → Microphone
- Enable for Terminal/iTerm

**Solution** (Linux): Add user to audio group
```bash
sudo usermod -a -G audio $USER
# Log out and back in
```

## Next Steps

- **[Configuration Guide](CONFIGURATION.md)** - Advanced settings and tuning
- **[API Reference](API.md)** - Complete HTTP API documentation
- **[Deployment Guide](DEPLOYMENT.md)** - Production k8s/k3s deployment
- **[Architecture](ARCHITECTURE.md)** - Code structure and design patterns
- **[Troubleshooting](TROUBLESHOOTING.md)** - Detailed problem-solving guide
