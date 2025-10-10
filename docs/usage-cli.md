# CLI Usage

Build
- `make deps` (libwhisper)
- `go build -tags "cli whisper" -o dist/gosper ./cmd/gosper`
- For mic + beep: `-tags "cli malgo whisper"`

Commands
- `transcribe <file>`
  - Flags: `--model`, `--lang`, `--translate`, `--threads`, `--timestamps`, `--beam`, `--max-tokens`, `--prompt`, `-o/--out`
- `record`
  - Flags: `--device`, `--duration`, `--model`, `--lang`, `-o/--out`, `--audio-feedback`, `--output-device`, `--beep-volume`
- `devices list` — show inputs; marks current with `*`
- `devices select <id|name>` — persist default
- `version`

Examples
- `dist/gosper transcribe whisper.cpp/samples/jfk.wav --model /abs/path/ggml-tiny.en.bin --lang en -o out.txt`
- `dist/gosper devices list`
- `dist/gosper devices select "External USB Mic"`
- `dist/gosper record --duration 5s --audio-feedback --output-device index:1 --beep-volume 0.3`

Config persistence
- `~/.config/gosper/config.json` holds: LastDeviceID, AudioFeedback, OutputDeviceID, BeepVolume
- CLI uses file as defaults and writes changes after commands

