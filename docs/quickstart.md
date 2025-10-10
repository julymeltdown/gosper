# Quickstart

This guide shows copy-paste commands for the most common flows.

Prereqs
- Go 1.22+
- A C toolchain (for libwhisper/malgo builds)

Setup
```
make deps               # builds whisper.cpp static lib
```

CLI builds
```
# CLI + inference (no mic)
go build -tags "cli whisper" -o dist/gosper ./cmd/gosper

# CLI + mic + inference (malgo)
go build -tags "cli malgo whisper" -o dist/gosper ./cmd/gosper
```

Transcribe a file
```
dist/gosper transcribe whisper.cpp/samples/jfk.wav \
  --model /absolute/path/to/ggml-tiny.en.bin --lang en -o out.txt
```

List/select devices (malgo)
```
dist/gosper devices list
dist/gosper devices select "External USB Mic"
```

Record with beep (5 seconds)
```
dist/gosper record --duration 5s --audio-feedback \
  --output-device index:1 --beep-volume 0.3 \
  --model /absolute/path/to/ggml-tiny.en.bin -o rec.txt
```

Run unit tests
```
make test
```

Run integration tests (requires model file)
```
export GOSPER_MODEL_PATH=/absolute/path/to/ggml-tiny.en.bin
GOSPER_INTEGRATION=1 go test ./test/integration -tags whisper -v
```

Server (Docker)
```
docker build -f Dockerfile.server -t handy/server:local .
docker run -p 8080:8080 handy/server:local

curl -F audio=@whisper.cpp/samples/jfk.wav \
  -F model=ggml-tiny.en.bin -F lang=en \
  http://localhost:8080/api/transcribe
```

k3s Deploy
```
docker build -f Dockerfile.server -t handy/server:local .
docker build -f Dockerfile.frontend -t handy/fe:local .
cp scripts/k3s/env.example scripts/k3s/.env
sed -i'' -e 's/handy.local/handy.local/g' scripts/k3s/.env
bash scripts/k3s/deploy.sh
# Add to /etc/hosts if needed: 127.0.0.1 handy.local
open http://handy.local/
```

