# HTTP API

Server binary
- Built from `cmd/server` (Go), depends on whisper adapter (`-tags=whisper`)
- Endpoint: `POST /api/transcribe`

POST /api/transcribe
- Content-Type: `multipart/form-data`
- Fields:
  - `audio` — audio file (wav recommended)
  - `model` — model name or local path (optional; defaults from env)
  - `lang` — language or `auto` (optional)

Response 200 application/json
```
{
  "language": "en",
  "text": "... full transcript ...",
  "segments": [
    {"Index":0, "StartMS":0, "EndMS":2100, "Text":"..."}
  ],
  "duration_ms": 1234
}
```

Errors
- 400 — bad request (missing file)
- 502 — transcription failure

Env
- `PORT` (default 8080)
- `GOSPER_MODEL` (`ggml-tiny.en.bin` default)
- `GOSPER_LANG` (`auto` default)
- `MODEL_BASE_URL` (download base for models if not local)

