# Models

## Model Sources

**Official Source**: Hugging Face
```
https://huggingface.co/ggerganov/whisper.cpp/resolve/main/
```

**Available Models**:
- English-only: `ggml-tiny.en.bin`, `ggml-base.en.bin`, `ggml-small.en.bin`, `ggml-medium.en.bin`
- Multilingual: `ggml-tiny.bin`, `ggml-base.bin`, `ggml-small.bin`, `ggml-medium.bin`, `ggml-large-v3.bin`

## Model Management

**Behavior**:
1. If model path is absolute → use directly
2. If model found in cache → use cached version
3. Otherwise → download from `MODEL_BASE_URL`
4. Verify SHA256 checksum (optional)
5. Retry with exponential backoff on failure

**Cache Directory**:
- Linux: `~/.cache/gosper/`
- macOS: `~/Library/Caches/gosper/`
- Windows: `%LOCALAPPDATA%\gosper\cache\`

**Custom Cache**:
```bash
export GOSPER_CACHE=/path/to/cache
```

**Custom Source**:
```bash
export MODEL_BASE_URL=https://your-cdn.com/models/
```

## Default Model

**Default**: `ggml-tiny.en.bin` (fast, English-only)

For better accuracy, use larger models:
- `ggml-base.en.bin` - Balanced speed/accuracy
- `ggml-medium.en.bin` - High accuracy
- `ggml-large-v3.bin` - Maximum accuracy (multilingual)

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
- **Bitrate**: All bitrates (CBR, VBR, ABR)
- **File Size**: Maximum 200 MB

**Processing**:
- All audio is resampled to 16 kHz (Whisper requirement)
- Stereo audio is downmixed to mono
- Samples are normalized to float32 [-1, 1]

For detailed configuration options, see [CONFIGURATION.md](CONFIGURATION.md).
