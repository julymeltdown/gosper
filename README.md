# Gosper

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](go.mod)

**Self-hosted speech-to-text API service powered by OpenAI Whisper**

Gosper is a production-ready transcription service that runs entirely on your infrastructure. Deploy it to k3s, send audio files via HTTP API, and get accurate transcripts back—all without your users' voice data ever touching a cloud provider's servers.

## Why Gosper?

Building an app with speech-to-text features? You have options:

| Your Choice | Gosper | Cloud APIs (Google, AWS, Azure) |
|------------|--------|--------------------------------|
| **Privacy** | ✅ Audio stays on your servers | ❌ Audio sent to cloud |
| **Cost** | ✅ Free after deployment | ❌ $$ per minute |
| **Control** | ✅ You own the infrastructure | ❌ Vendor lock-in |
| **Accuracy** | ✅ OpenAI Whisper | ✅ High accuracy |

**Gosper was built for developers who:**
- 🔒 Care about user privacy and data sovereignty
- 💰 Want to avoid escalating per-minute API costs
- 🏗️ Prefer self-hosted infrastructure (homelab, VPS, on-prem)
- 🚀 Need a production-ready backend for mobile/web apps
- 🛠️ Value clean, extensible architecture

*"Your users' voices shouldn't be a subscription service."*

## How It Works

### For App Developers
1. **Deploy** Gosper to your k3s cluster or Docker host
2. **Integrate** your mobile/web app with the `/api/transcribe` endpoint
3. **Send** audio files (WAV or MP3) via HTTP POST
4. **Receive** accurate JSON transcripts powered by Whisper
5. **Scale** with your userbase—no per-minute costs

### For CLI Users
1. **Build** the CLI binary using `make build-all`.
2. **Download a Model**: Gosper needs a Whisper model to run.
   ```bash
   # Build the model downloader utility
   make -C whisper.cpp/bindings/go examples

   # Download the tiny English model
   ./whisper.cpp/bindings/go/build_go/go-model-download -out whisper.cpp/models ggml-tiny.en.bin
   ```
3. **Transcribe an Audio File**:
   ```bash
   # Transcribe a WAV file
   ./dist/gosper transcribe path/to/your/audio.wav --model whisper.cpp/models/ggml-tiny.en.bin
   ```
   **Note**: The current version has a known issue with MP3 decoding. Please use WAV files for transcription.

All processing happens locally using [whisper.cpp](https://github.com/ggerganov/whisper.cpp), a high-performance C++ implementation of OpenAI's Whisper model.

## Quick Start

### Try It in 30 Seconds (Docker)

**Note**: The public Docker image `gosper/server:latest` is currently out of date. Please build the image locally.

```bash
# Build the server image
docker build -f Dockerfile.server -t gosper/server:local .

# Run the service
docker run -p 8080:8080 gosper/server:local

# Transcribe an audio file
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@your-audio.mp3" \
  -F "lang=auto"
```

🎉 **That's it!** Your transcript is returned as JSON.

### Quick Start Guides
- 📦 **[Docker & Docker Compose](docs/QUICKSTART.md#docker-quick-start)** - Run locally in seconds
- ☸️ **[Kubernetes/k3s Deployment](docs/deployment-complete.md)** - Production setup
- 💻 **[CLI Installation](docs/QUICKSTART.md#cli-quick-start)** - Command-line usage
- 🛠️ **[Build from Source](docs/BUILD.md)** - Development setup

## Features

- 🎙️ **Multiple Interfaces**: HTTP API, CLI, and Web UI
- 🎵 **Format Support**: WAV and MP3 with automatic detection
- 🌍 **Multi-Language**: 100+ languages with auto-detection
- ⚡ **Fast**: Optimized whisper.cpp with parallelization
- 🐳 **Production-Ready**: Docker images and k8s manifests included
- 🏗️ **Clean Architecture**: Hexagonal design, 85%+ test coverage
- 📴 **Offline Capable**: Models cached locally, no internet required

## Architecture

Gosper follows hexagonal (ports & adapters) architecture:

```
┌─────────────────────────────────────────────┐
│        Inbound Adapters                     │
│     (HTTP API, CLI, Web UI)                 │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│         Use Cases                           │
│  (TranscribeFile, RecordAndTranscribe)      │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│        Outbound Adapters                    │
│  (Whisper.cpp, Audio Decoders, Storage)    │
└─────────────────────────────────────────────┘
```

📚 **[Full Architecture Guide](docs/ARCHITECTURE.md)** - Detailed layer descriptions and extension points

## Platform Support

- ✅ **Linux** (x86_64, ARM64) - Ubuntu 20.04+, Debian 11+
- ✅ **macOS** (Intel, Apple Silicon) - macOS 11+
- ✅ **Windows** (x86_64) - Windows 10+
- ✅ **Docker** - Multi-platform images available
- ✅ **Kubernetes** - k3s/k8s manifests and Helm charts

🔧 **[Platform-Specific Notes](docs/TROUBLESHOOTING.md#platform-specific-issues)** - Build requirements and known issues

## Documentation

### Getting Started
- 🚀 **[Quick Start Guide](docs/QUICKSTART.md)** - Get transcribing in minutes
- ☸️ **[Deployment Guide](docs/deployment-complete.md)** - Production k3s/k8s setup
- 🛠️ **[Build from Source](docs/BUILD.md)** - Development environment

### Reference
- 🏗️ **[Architecture](docs/ARCHITECTURE.md)** - Design principles and code structure
- 🔌 **[API Reference](docs/API.md)** - HTTP API endpoints and examples
- ⚙️ **[Configuration](docs/CONFIGURATION.md)** - Environment variables and models
- 🩺 **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

### Contributing
- 🤝 **[Contributing Guide](docs/CONTRIBUTING.md)** - Development workflow and guidelines

## How to Contribute

We welcome contributions! Gosper aims to be not just useful, but also forkable and extensible.

1. **Check existing issues** at [github.com/cjpais/go-whisper/issues](https://github.com/cjpais/go-whisper/issues)
2. **Read the contributing guide** at [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)
3. **Fork and create a feature branch**
4. **Write tests** - We maintain 85%+ coverage
5. **Submit a pull request** with clear description

### Development Setup
```bash
# Clone and build
git clone https://github.com/yourusername/gosper.git
cd gosper

# Build all binaries
make build-all

# Run tests
make test
```

See [docs/BUILD.md](docs/BUILD.md) for detailed setup instructions.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Gosper stands on the shoulders of giants:

- **[OpenAI Whisper](https://github.com/openai/whisper)** - Revolutionary speech recognition model
- **[whisper.cpp](https://github.com/ggerganov/whisper.cpp)** - High-performance C++ implementation
- **[hajimehoshi/go-mp3](https://github.com/hajimehoshi/go-mp3)** - Pure Go MP3 decoder
- **[Go Community](https://golang.org)** - Excellent language and ecosystem

---

*"Self-host your speech-to-text. Own your data. Build without limits."*
