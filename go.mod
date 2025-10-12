module gosper

go 1.23

toolchain go1.24.3

require (
	github.com/gen2brain/malgo v0.11.24
	github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-00010101000000-000000000000
	github.com/hajimehoshi/go-mp3 v0.3.4
	github.com/spf13/cobra v1.8.1 // used under build tag 'cli'
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/ggerganov/whisper.cpp/bindings/go => ./whisper.cpp/bindings/go
