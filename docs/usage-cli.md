# CLI Usage

## Build

The recommended way to build the CLI is using the main Makefile from the root of the repository.

```bash
# Build CLI, server, and all dependencies
make build-all
```

This will place the `gosper` binary in the `dist/` directory.

## Getting Started: First Transcription

**1. Download a Model**

Gosper requires a Whisper model file to perform transcription.
```bash
# Build the model downloader utility
make -C whisper.cpp/bindings/go examples

# Download the tiny English model
./whisper.cpp/bindings/go/build_go/go-model-download -out whisper.cpp/models ggml-tiny.en.bin
```

**2. Transcribe a WAV File**

Use the `transcribe` command to process an audio file.

> **Note**: We strongly recommend using WAV files. The current MP3 decoder has a bug that may cause transcription to fail.

```bash
./dist/gosper transcribe whisper.cpp/samples/jfk.wav \
  --model whisper.cpp/models/ggml-tiny.en.bin
```

## Commands

### `transcribe <file>`
Transcribes an audio file.

- **Flags**:
  - `--model <path>`: Path to the Whisper model file (required).
  - `--lang <language>`: Spoken language in the audio (`en`, `es`, `auto`, etc.). Default: `auto`.
  - `--out <filepath>`: Path to save the transcript (e.g., `transcript.txt`).
  - `--threads <num>`: Number of CPU threads to use.
  - (See `--help` for all flags)

### `record`
Records audio from a microphone and transcribes it.

- **Flags**:
  - `--duration <time>`: Recording duration (e.g., `30s`, `1m`).
  - `--device <id>`: ID of the recording device to use.
  - (See `--help` for all flags)

### `devices`
Manage audio input devices.

- `devices list`: Show available recording devices.
- `devices select <id>`: Set the default recording device.

### `version`
Show the application version.

## Examples

**Transcribe and Save to File**
```bash
./dist/gosper transcribe my_audio.wav \
  --model whisper.cpp/models/ggml-tiny.en.bin \
  --lang en \
  -o my_audio.txt
```

**List Recording Devices**
```bash
./dist/gosper devices list
```

**Record for 10 Seconds**
```bash
./dist/gosper record --duration 10s --model whisper.cpp/models/ggml-tiny.en.bin
```

Config persistence
- `~/.config/gosper/config.json` holds: LastDeviceID, AudioFeedback, OutputDeviceID, BeepVolume
- CLI uses file as defaults and writes changes after commands

