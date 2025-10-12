# Dependencies

This document outlines the dependencies of the Gosper project.

## Go Modules

The following Go modules are used in this project:

- **github.com/gen2brain/malgo**: Used for audio recording.
- **github.com/hajimehoshi/go-mp3**: Used for decoding MP3 files.
- **github.com/spf13/cobra**: Used for creating the command-line interface.
- **github.com/stretchr/testify**: Used for assertions in tests.

## CGO Dependencies

The following C libraries are used in this project:

- **whisper.cpp**: Used for speech-to-text transcription. This is included as a git submodule.
- **miniaudio**: Used by `malgo` for audio recording. This is included as part of the `malgo` library.
