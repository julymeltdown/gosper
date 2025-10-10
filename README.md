# Gosper

CLI for local speech-to-text using whisper.cpp via Go bindings, implemented with hexagonal architecture.

📘 **[Complete Deployment Guide](./docs/deployment-complete.md)** — Covers Docker Compose, Kubernetes, AWS/GCP/Azure, CI/CD, monitoring, and more with detailed Mermaid diagrams.

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

## k3s Deployment

This repo includes a minimal backend HTTP server and static frontend for running in k3s.

Images
- Backend: `Dockerfile.server` builds a Linux binary linking against `whisper.cpp` and exposes `/api/transcribe`.
- Frontend: `Dockerfile.frontend` serves `web/` via nginx.

Kubernetes Manifests: `deploy/k8s/base/*.yaml`
- Namespace, backend Deployment/Service, frontend Deployment/Service, Ingress (Traefik / k3s default)

Quickstart
1) Build images (or set BE_IMAGE/FE_IMAGE to your registry/tag):
   - `docker build -f Dockerfile.server -t gosper/server:local .`
   - `docker build -f Dockerfile.frontend -t gosper/fe:local .`
2) Configure env:
   - `cp scripts/k3s/env.example scripts/k3s/.env` and edit BE_IMAGE/FE_IMAGE, NAMESPACE, DOMAIN
3) Deploy:
   - `bash scripts/k3s/deploy.sh`
4) Open `http://$DOMAIN/` and upload a WAV file

Notes
- For production, push images to a registry (Docker Hub, GHCR) and set BE_IMAGE/FE_IMAGE accordingly in `.env`.
- k3s includes Traefik; the provided Ingress uses `web` entrypoint. Configure DNS or add `/etc/hosts` for `${DOMAIN}`.
- The server fetches models from Hugging Face on first use; ensure outbound network is allowed.

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
