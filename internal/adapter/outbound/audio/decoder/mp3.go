package decoder

import "fmt"

type mp3Decoder struct{}

func NewMP3(path string) (Decoder, error) {
    return nil, fmt.Errorf("mp3 decoding not implemented; please convert to wav")
}

