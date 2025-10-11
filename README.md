# Gosper

**Privacy-first speech-to-text service powered by OpenAI Whisper, running entirely on your infrastructure.**

Gosper is a comprehensive speech-to-text solution that converts audio files and live microphone recordings into accurate text transcripts. Unlike cloud-based alternatives, Gosper runs OpenAI's Whisper model locally, ensuring your audio data never leaves your infrastructure.

## Table of Contents

- [What is Gosper?](#what-is-gosper)
- [Key Features](#key-features)
- [Use Cases](#use-cases)
- [Why Choose Gosper?](#why-choose-gosper)
- [Quick Demo](#quick-demo)
- [Architecture Overview](#architecture-overview)
- [Build](#build)
- [Production Deployment (k3s/Kubernetes)](#production-deployment-k3skubernetes)
- [CLI Quickstart](#quickstart-cli)
- [Build Tags](#build-tags)
- [Configuration](#config--persistence)
- [Testing](#testing--coverage)
- [Troubleshooting](#troubleshooting)

## What is Gosper?

Gosper provides two ways to transcribe speech:

1. **CLI Tool** — Command-line interface for batch processing audio files and recording from microphone
2. **Web Service** — HTTP API with browser-based UI for easy file uploads and transcription

All processing happens locally using [whisper.cpp](https://github.com/ggerganov/whisper.cpp), a high-performance C++ implementation of OpenAI's Whisper automatic speech recognition model.

## Key Features

- **🔒 Privacy-First**: All transcription happens locally—no data sent to external APIs
- **🌍 Multi-Language**: Supports 100+ languages with automatic language detection
- **💰 Cost-Effective**: No per-minute API costs; runs on your own hardware
- **📴 Offline Capable**: Works without internet connection (models cached locally)
- **🎤 Live Recording**: Record and transcribe directly from microphone
- **⚡ Fast Processing**: Optimized C++ implementation with OpenMP parallelization
- **🐳 Cloud-Ready**: Docker images and Kubernetes manifests for easy deployment
- **🏗️ Clean Architecture**: Hexagonal architecture with dependency injection for testability

## Use Cases

- **Meeting Transcription**: Record and transcribe team meetings, interviews, or calls
- **Content Creation**: Generate subtitles for podcasts, videos, or presentations
- **Documentation**: Convert voice notes and recordings into searchable text
- **Accessibility**: Create text alternatives for audio content
- **Research**: Transcribe interviews, focus groups, or field recordings

## Why Choose Gosper?

| Feature | Gosper | Cloud APIs (e.g., Google, AWS) |
|---------|--------|--------------------------------|
| **Data Privacy** | ✅ Runs locally | ❌ Data sent to cloud |
| **Cost** | ✅ Free after setup | ❌ Pay per minute |
| **Offline Use** | ✅ Works offline | ❌ Requires internet |
| **Accuracy** | ✅ OpenAI Whisper | ✅ High accuracy |
| **Deployment** | ✅ Self-hosted | ❌ Vendor lock-in |

## Quick Demo

### CLI Usage
```bash
# Transcribe an audio file
./gosper transcribe meeting.wav --model ggml-base.en.bin --lang en

# Record 30 seconds from microphone and transcribe
./gosper record --duration 30s --audio-feedback
```

### Web API
```bash
# Start the server
docker run -p 8080:8080 gosper/server:local

# Transcribe via API
curl -F audio=@meeting.wav http://localhost:8080/api/transcribe
```

### Web UI
Upload audio files through a simple browser interface at `http://localhost:8080/`

## Architecture Overview

Gosper follows hexagonal (ports & adapters) architecture for clean separation of concerns:

```
┌─────────────────────────────────────────────┐
│           Inbound Adapters                  │
│  (CLI, HTTP Server, Web UI)                 │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│         Use Cases (Business Logic)          │
│  (TranscribeFile, RecordAndTranscribe)      │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│        Outbound Adapters                    │
│  (Whisper.cpp, Audio I/O, Model Fetcher)   │
└─────────────────────────────────────────────┘
```

---

📘 **[Complete Deployment Guide](./docs/deployment-complete.md)** — Detailed guide covering Docker Compose, Kubernetes, AWS/GCP/Azure, CI/CD, monitoring, and more.

## Build

1. Build whisper static library:

```
make deps
```

2. Build the CLI binary:

```
make build
```

3. Run tests:

```
make test
```

## Environment

- `GOSPER_MODEL`: model name or path (default `ggml-tiny.en.bin`)
- `GOSPER_LANG`: language code or `auto` (default `auto`)
- `GOSPER_THREADS`: integer thread count
- `GOSPER_CACHE`: model cache directory (defaults to OS cache dir)
- `GOSPER_LOG`: log level (debug|info|warn|error)

## Notes

- This project uses whisper.cpp Go bindings under `whisper.cpp/bindings/go`.
- To enable integration tests, set `GOSPER_INTEGRATION=1` and run `make itest`.

## Platform Notes

### macOS
- The first time you run `gosper record`, macOS prompts for Microphone access. Approve it under System Settings → Privacy & Security → Microphone.
- If the prompt doesn’t appear, run from Terminal and check System Settings for denied permissions.

### Linux
- Audio capture may require PulseAudio or PipeWire. Ensure your user is in the appropriate audio groups.
- If devices are not found, try running inside an environment with Pulse/ALSA configured.

### Windows
- The `malgo` input adapter builds under the `malgo` build tag. Ensure you have a C compiler and run from a terminal with microphone permission enabled in Settings.

## Production Deployment (k3s/Kubernetes)

Gosper includes production-ready Docker images and Kubernetes manifests for deploying both the backend API and frontend web UI.

### What Gets Deployed

- **Backend Service**: Go HTTP server exposing `/api/transcribe` endpoint
- **Frontend Service**: Nginx serving static web UI for file uploads
- **Model Auto-Download**: Whisper models are fetched from Hugging Face on first use
- **NodePort Services**: Direct access via node IP and port (configurable)
- **Ingress**: Traefik-based routing for domain-based access

### Quick Deployment

```bash
# 1. Build Docker images
docker build -f Dockerfile.server -t gosper/server:local .
docker build -f Dockerfile.frontend -t gosper/fe:local .

# 2. Import images to k3s
docker save gosper/server:local -o /tmp/gosper-server.tar
docker save gosper/fe:local -o /tmp/gosper-fe.tar
sudo k3s ctr images import /tmp/gosper-server.tar
sudo k3s ctr images import /tmp/gosper-fe.tar

# 3. Configure deployment (optional - defaults provided)
cp scripts/k3s/env.example scripts/k3s/.env
# Edit: NAMESPACE, BE_NODEPORT, FE_NODEPORT, DOMAIN

# 4. Deploy to k3s
bash scripts/k3s/deploy.sh
```

### Access Your Deployment

After deployment, Gosper is accessible via:

- **Backend API**: `http://<NODE_IP>:<BE_NODEPORT>/api/transcribe`
- **Frontend UI**: `http://<NODE_IP>:<FE_NODEPORT>/`
- **Ingress** (if DNS configured): `http://<DOMAIN>/`

Example with default ports (31209 for backend, 31987 for frontend):
```bash
# Upload and transcribe via API
curl -F audio=@meeting.wav http://192.168.1.100:31209/api/transcribe

# Access web UI
open http://192.168.1.100:31987
```

### Configuration Options

Edit `scripts/k3s/.env` to customize:

```bash
# Namespace and domain
export NAMESPACE=gosper
export DOMAIN=gosper.local

# Docker images (use registry URLs for production)
export BE_IMAGE=gosper/server:local
export FE_IMAGE=gosper/fe:local

# NodePort assignments (must be in range 30000-32767)
export BE_NODEPORT=31209
export FE_NODEPORT=31987
```

### Production Considerations

- **Image Registry**: Push images to GHCR, Docker Hub, or private registry for production
- **Resource Limits**: Adjust CPU/memory requests in `deploy/k8s/base/*.yaml` based on workload
- **Model Size**: Larger Whisper models (medium, large) require more memory
- **TLS/HTTPS**: Configure cert-manager and update Ingress annotations for SSL
- **Monitoring**: See [deployment guide](./docs/deployment-complete.md) for Prometheus/Grafana setup

### Verify Deployment

```bash
# Check pod status
kubectl get pods -n gosper

# Check services and ports
kubectl get svc -n gosper

# View backend logs
kubectl logs -f deployment/gosper-be -n gosper
```

## Quickstart (CLI)

- Build whisper static lib: `make deps`
- Build CLI (inference): `go build -tags "cli whisper" -o dist/gosper ./cmd/gosper`
- Transcribe a file: `dist/gosper transcribe whisper.cpp/samples/jfk.wav --model /abs/path/ggml-tiny.en.bin --lang en -o out.txt`
- List devices (malgo): `go build -tags "cli malgo" -o dist/gosper ./cmd/gosper && dist/gosper devices list`
- Record 5s with beep: `go build -tags "cli malgo whisper" -o dist/gosper ./cmd/gosper && dist/gosper record --duration 5s --audio-feedback`

## Build Tags

- `cli`: include CLI commands and entrypoint
- `whisper`: enable whisper inference adapters and server
- `malgo`: enable mic capture and output beep via miniaudio

Combine tags per need, e.g. CLI + mic + inference: `-tags "cli malgo whisper"`.

## Architecture & Mechanism

- Ports (`internal/port`): abstractions for Transcriber, AudioInput, ModelRepo, Storage, Logger, Clock
- Use cases (`internal/usecase`): orchestrate pure flows (TranscribeFile, RecordAndTranscribe, ListDevices)
- Adapters (`internal/adapter`):
  - Inbound CLI (Cobra) behind `cli`
  - Outbound Whisper behind `whisper` using `whisper.cpp/bindings/go/pkg/whisper`
  - Outbound Audio behind `malgo` using `github.com/gen2brain/malgo` (capture/playback)
  - Outbound Model (cache+download+checksum), Storage (atomic txt/json)

Flow
- File: decode WAV (PCM16/float32) → normalize/downmix → resample 16k mono → Whisper `Process` → collect segments → write transcript
- Mic: malgo capture (f32) → downmix → (16k) buffer until stop → Whisper `Process` → write transcript
- Device search: exact id → exact name (ci) → prefix → substring → fuzzy (Levenshtein) → persist selection in config

## Config & Persistence

Environment (defaults)
- `GOSPER_MODEL`, `GOSPER_LANG`, `GOSPER_THREADS`, `GOSPER_CACHE`, `GOSPER_LOG`
- `GOSPER_AUDIO_FEEDBACK=1`, `GOSPER_OUTPUT_DEVICE`, `GOSPER_BEEP_VOLUME`

Config file: `~/.config/gosper/config.json`
- Persists `LastDeviceID`, `AudioFeedback`, `OutputDeviceID`, `BeepVolume`
- CLI reads defaults and updates after use

## Models & Checksums
- Resolves paths, caches under OS cache, optional download from `MODEL_BASE_URL`
- Optional sha256 verification and retry with backoff during download

## Testing & Coverage

- Unit tests: `make test` (fakes for ports; deterministic)
- Integration: `GOSPER_INTEGRATION=1 go test ./test/integration -tags whisper -v`
- CI gates: total ≥ 85%, usecase package ≥ 90%

## Server & Frontend

- Backend (Go): `cmd/server`, Dockerfile `Dockerfile.server`; exposes `/api/transcribe`
- Frontend (static): `web/`, Dockerfile `Dockerfile.frontend`
- Example:
  - `docker build -f Dockerfile.server -t gosper/server:local . && docker run -p 8080:8080 gosper/server:local`
  - `curl -F audio=@whisper.cpp/samples/jfk.wav -F model=ggml-tiny.en.bin -F lang=en http://localhost:8080/api/transcribe`

## Troubleshooting

- Mic permissions (macOS): approve in System Settings → Privacy & Security → Microphone
- Linux audio: ensure Pulse/PipeWire/ALSA present; try alternate device ids
- CGO build errors: run `make deps`; ensure `C_INCLUDE_PATH` and `LIBRARY_PATH` point to `whisper.cpp` and `bindings/go/build`
- Integration flakes: supply `GOSPER_MODEL_PATH` to an existing local model

## Roadmap

- Always-on mode with VAD segmentation
- Higher-fidelity resampler (sinc) behind build tag
- Output device management commands; volume profiles
- Helm chart and GH Actions CI/CD to k3s
