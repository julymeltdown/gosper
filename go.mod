module gosper

go 1.22

require (
    github.com/hajimehoshi/go-mp3 v0.3.4
    github.com/spf13/cobra v1.8.1 // indirect; used under build tag 'cli'
)

replace github.com/ggerganov/whisper.cpp/bindings/go => ./whisper.cpp/bindings/go
