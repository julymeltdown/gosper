# Configuration

Precedence
1) CLI flags (commands)
2) Config file (`~/.config/gosper/config.json`)
3) Environment variables
4) Built-in defaults

Env variables
- `GOSPER_MODEL` — model name/path (default `ggml-tiny.en.bin`)
- `GOSPER_LANG` — language or `auto`
- `GOSPER_THREADS` — integer threads
- `GOSPER_CACHE` — model cache dir
- `GOSPER_LOG` — log level (debug|info|warn|error)
- `GOSPER_AUDIO_FEEDBACK` — `1` to enable beeps
- `GOSPER_OUTPUT_DEVICE` — output device id/name for beep
- `GOSPER_BEEP_VOLUME` — 0..1

Config file schema
```json
{
  "Model": "ggml-tiny.en.bin",
  "Language": "auto",
  "Threads": 0,
  "CacheDir": "",
  "LogLevel": "info",
  "LastDeviceID": "index:0",
  "AudioFeedback": true,
  "OutputDeviceID": "index:1",
  "BeepVolume": 0.3
}
```

