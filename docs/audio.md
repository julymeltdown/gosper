# Audio Pipeline

Decoders
- WAV (PCM16/float32) supported now; MP3 stubbed
- Normalize to [-1,1], downmix stereo → mono

Resampler
- Linear interpolation to 16kHz mono
- Unit tests for identity, constant, up/downsample, short sequences
- Future: sinc resampler behind build tag

Microphone capture
- `malgo` build tag enables capture via miniaudio
- Frames delivered as f32 mono via channel; downmix and resample handled by adapter

Beeps (feedback)
- Console bell (default) or malgo playback (output device + volume)
- CLI flags: `--audio-feedback`, `--output-device`, `--beep-volume`

