# Build Tags

`cli`
- Includes CLI commands and entrypoint

`whisper`
- Enables whisper adapter and server; requires libwhisper static lib

`malgo`
- Enables mic capture and output beeps via miniaudio

Combos
- CLI + inference: `-tags "cli whisper"`
- CLI + mic + inference: `-tags "cli malgo whisper"`
- Server: `-tags whisper`

