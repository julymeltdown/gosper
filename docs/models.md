# Models

Sources
- Hugging Face: `https://huggingface.co/ggerganov/whisper.cpp/resolve/main/models/`

Behavior
- If provided a path, use it directly
- Else check cache dir (OS default or `GOSPER_CACHE`)
- Else download from `MODEL_BASE_URL` with retries; optional sha256 verification

Defaults
- `ggml-tiny.en.bin` for speed; pass larger models for better accuracy

